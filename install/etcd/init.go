package etcd

import (
	"fmt"
	etcdcerts "github.com/ihaiker/vik8s/certs/etcd"
	"github.com/ihaiker/vik8s/install"
	"github.com/ihaiker/vik8s/install/repo"
	"github.com/ihaiker/vik8s/install/tools"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"os"
	"path/filepath"
)

func InitCluster(node *ssh.Node) {
	install.PreCheck(node)
	checkEtcdadm(node)
	makeAndPushCerts(node)
	etcdadmInit(node)
	Config.Join(node.Host)
}

func makeAndPushCerts(node *ssh.Node) {
	node.Logger("make certs files")
	name := node.Hostname
	dir := tools.Join("etcd", "pki")
	sans := []string{"127.0.0.1", "localhost", node.Hostname, node.Host}
	sans = append(sans, utils.ParseIPS(Config.Nodes)...)
	sans = append(sans, Config.ServerCertExtraSans...)
	vt := Config.CertsValidity
	etcdcerts.CreatePKIAssets(name, dir, sans, vt)

	certsFiles := map[string]string{
		"ca":                    "ca",
		"server-" + name:        "server",
		"peer-" + name:          "peer",
		"etcdctl-etcd-client":   "etcdctl-etcd-client",
		"apiserver-etcd-client": "apiserver-etcd-client",
		"healthcheck-client":    "healthcheck-client",
	}

	for lf, rf := range certsFiles {
		for _, exp := range []string{".key", ".crt"} {
			local := filepath.Join(dir, lf+exp)
			remote := filepath.Join(Config.CertsDir, rf+exp)
			utils.Panic(node.Scp(local, remote), "scp %s %s", local, remote)
		}
	}
}

func checkEtcdadm(node *ssh.Node) {
	utils.Line("check and install etcdadm")

	etcdadm, err := node.Cmd2String("command -v etcdadm")
	if err != nil {
		etcdadm = installEtcdadm(node)
	}
	//etcdadm
	{
		local := tools.Join("etcd", "etcdadm")
		if utils.NotExists(local) {
			err = node.Pull(etcdadm, local)
			utils.Panic(err, "pull etcdadm")
		}
	}
	//etcd.tar.gz
	{
		tar := fmt.Sprintf("etcd-v%s-linux-amd64.tar.gz", Config.Version)
		local := tools.Join("etcd", tar)
		if utils.Exists(local) {
			remote := fmt.Sprintf("/var/cache/etcdadm/etcd/v%s/%s", Config.Version, tar)
			err := node.Scp(local, remote)
			utils.Panic(err, "scp %s %s", local, remote)
		}
	}
}

func installEtcdadm(node *ssh.Node) string {
	remoteBin := "/usr/local/bin/etcdadm"
	localBin := tools.Join("etcd", "etcdadm")

	if utils.NotExists(localBin) {
		node.Logger("build etcdadm")
		tools.Install("git", "", node)
		tools.Install("golang", "", node)
		source := Config.Source
		if source == "" {
			source = repo.Etcdadm()
		}
		goProxy := ""
		if tools.China {
			goProxy = `export GOPROXY="https://goproxy.io"`
		}
		shell := fmt.Sprintf(`
cd /tmp
git clone %s --local etcdadm
cd etcdadm
%s
go build
mv -f etcdadm %s`, source, goProxy, remoteBin)
		err := node.ShellChannel(shell, utils.Stdout(node.Hostname))
		utils.Panic(err, "make etcdadm")
	} else {
		err := node.ScpProgress(localBin, remoteBin)
		utils.Panic(err, "scp etcdadm")
	}
	return remoteBin
}

func etcdadmInit(master *ssh.Node) {
	utils.Line("etcdadm init")
	cmd := "etcdadm init --name " + master.Hostname +
		" --install-dir /usr/local/bin " +
		" --certs-dir " + Config.CertsDir +
		" --version " + Config.Version
	if Config.Snapshot != "" {
		cmd += " --snapshot " + Config.Snapshot
	}
	/* use certs make
	for _, san := range Config.ServerCertExtraSans {
		cmd += " --server-cert-extra-sans " + san
	}
	*/
	err := master.CmdStd(cmd, os.Stdout)
	utils.Panic(err, "etcdadm init")

	tar := fmt.Sprintf("etcd-v%s-linux-amd64.tar.gz", Config.Version)
	local := tools.Join("etcd", tar)
	if utils.NotExists(local) {
		remote := fmt.Sprintf("/var/cache/etcdadm/etcd/v%s/%s", Config.Version, tar)
		err = master.Pull(remote, local)
		utils.Panic(err, "pull %s %s", remote, local)
	}
}
