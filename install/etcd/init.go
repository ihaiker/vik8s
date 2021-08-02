package etcd

import (
	"fmt"
	etcdcerts "github.com/ihaiker/vik8s/certs/etcd"
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/install/bases"
	"github.com/ihaiker/vik8s/install/cri"
	"github.com/ihaiker/vik8s/install/paths"
	"github.com/ihaiker/vik8s/install/repo"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"os"
	"path/filepath"
)

func InitCluster(node *ssh.Node) {
	node.Logger("install etcd server")
	bases.Check(node)
	cri.Install(node)
	pullContainerImage(node)
	makeAndPushCerts(node)
}

func pullContainerImage(node *ssh.Node) {
	var err error
	if config.Config.Docker != nil {
		dockerUrl := fmt.Sprintf("docker pull %s/%s:%s", repo.QuayIO(""), "coreos/etcd", config.Config.ETCD.Version)
		err = node.SudoCmdOutput(dockerUrl, os.Stdout)
	} else {
		err = node.SudoCmdOutput("ctr pull ", os.Stdout)
	}
	utils.Panic(err, "pull image")
}

func makeAndPushCerts(node *ssh.Node) {
	node.Logger("make certs files")

	name := node.Hostname
	dir := paths.Join("etcd", "pki")
	sans := []string{"127.0.0.1", "localhost", node.Hostname, node.Host}
	sans = append(sans, utils.ParseIPS(config.Config.ETCD.Nodes)...)
	sans = append(sans, config.Config.ETCD.ServerCertExtraSans...)
	vt := config.Config.ETCD.CertsValidity
	etcdcerts.CreatePKIAssets(name, dir, sans, vt)

	certsFiles := map[string]string{
		"ca":                    "ca",
		"server-" + name:        "server",
		"peer-" + name:          "peer",
		"etcdctl-etcd-client":   "etcdctl-etcd-client",
		"apiserver-etcd-client": "apiserver-etcd-client",
		"healthcheck-client":    "healthcheck-client",
	}

	for localFile, remoteFile := range certsFiles {
		for _, exp := range []string{".key", ".crt"} {
			local := filepath.Join(dir, localFile+exp)
			remote := filepath.Join(config.Config.ETCD.CertsDir, remoteFile+exp)
			utils.Panic(node.SudoScp(local, remote), "scp %s %s", local, remote)
		}
	}
}
