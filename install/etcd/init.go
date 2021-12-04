package etcd

import (
	"fmt"
	etcdcerts "github.com/ihaiker/vik8s/certs/etcd"
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/install/bases"
	"github.com/ihaiker/vik8s/install/cri"
	"github.com/ihaiker/vik8s/install/hosts"
	"github.com/ihaiker/vik8s/install/paths"
	"github.com/ihaiker/vik8s/install/repo"
	"github.com/ihaiker/vik8s/libs/logs"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func InitCluster(configure *config.Configuration, node *ssh.Node) {
	node.Logger("install etcd server")
	if configure.ETCD.Token == "" {
		token := utils.Random(16)
		node.Logger("make etcd token: %s", token)
		configure.ETCD.Token = token
	}
	bases.Check(node)
	cri.Install(configure, node)
	image := pullContainerImage(configure, node)
	cleanEtcdData(configure, node)
	makeAndPushCerts(configure, node)
	restoreSnapshot(configure, node, image)
	initEtcd(configure, node, image)
	waitEtcdReady(node)
	showClusterStatus(node)
}

func pullContainerImage(configure *config.Configuration, node *ssh.Node) (image string) {
	if configure.IsDockerCri() {
		repoUrl := configure.ETCD.Repo
		image = fmt.Sprintf("%s/%s:%s", repo.QuayIO(repoUrl), "coreos/etcd", configure.ETCD.Version)
		num, err := node.Sudo().CmdString(fmt.Sprintf("docker images --format '{{.Repository}}:{{.Tag}}' | grep %s | wc -l", image))
		utils.Panic(err, "check docker image tag")
		if num == "0" {
			err = node.Sudo().CmdOutput("docker pull "+image, os.Stdout)
			utils.Panic(err, "pull docker image")
		}
	}
	return
}

func makeAndPushCerts(configure *config.Configuration, node *ssh.Node) {
	node.Logger("make certs files")

	name := node.Hostname
	dir := CertsDir()
	sans := []string{"127.0.0.1", "localhost", node.Hostname, node.Host}
	sans = append(sans, configure.ETCD.ServerCertExtraSans...)
	sans = append(sans, configure.ETCD.Nodes...)
	vt := configure.ETCD.CertsValidity
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
			remote := filepath.Join(configure.ETCD.CertsDir, remoteFile+exp)
			utils.Panic(node.Sudo().Scp(local, remote), "scp %s %s", local, remote)
		}
	}
}

func initEtcd(configure *config.Configuration, node *ssh.Node, image string) {
	if configure.IsDockerCri() {
		initEtcdDocker(configure, node, image, "new")
	}
}

func initEtcdDocker(configure *config.Configuration, node *ssh.Node, image string, state string) {
	envs := map[string]string{
		"initial-advertise-peer-urls": "https://" + node.Host + ":2380",                        //对外通告该节点的同伴（Peer）监听地址，这个值会告诉集群中其他节点。
		"listen-peer-urls":            "https://" + node.Host + ":2380",                        //指定和 Cluster 其他 Node 通信的地址
		"listen-client-urls":          "https://" + node.Host + ":2379,https://127.0.0.1:2379", //指定对外提供服务的地址
		"advertise-client-urls":       "https://" + node.Host + ":2380",                        //对外通告的该节点的客户端监听地址，会告诉集群中其他节点。

		"initial-cluster-token": configure.ETCD.Token,            //创建集群
		"initial-cluster-state": state,                           //初始化新集群时使用 new, 加入已有集群时使用 existing
		"initial-cluster":       initialCluster(configure, node), //指定集群成员列表

		"client-cert-auth":      "true", //客户端 TLS 相关参数
		"trusted-ca-file":       "/etc/etcd/pki/ca.crt",
		"cert-file":             "/etc/etcd/pki/server.crt",
		"key-file":              "/etc/etcd/pki/server.key",
		"peer-client-cert-auth": "true", //集群内部 TLS 相关参数
		"peer-trusted-ca-file":  "/etc/etcd/pki/ca.crt",
		"peer-cert-file":        "/etc/etcd/pki/peer.crt",
		"peer-key-file":         "/etc/etcd/pki/peer.key",
	}
	ctlEnvs := map[string]string{
		"endpoints": "https://127.0.0.1:2379",
		"cacert":    "/etc/etcd/pki/ca.crt",
		"cert":      "/etc/etcd/pki/etcdctl-etcd-client.crt",
		"key":       "/etc/etcd/pki/etcdctl-etcd-client.key ",
	}
	cmd := "docker run -d --name vik8s-etcd --workdir /var/lib/etcd  --restart always --network host --hostname " + node.Hostname +
		" -v " + configure.ETCD.CertsDir + ":/etc/etcd/pki" +
		" -v " + configure.ETCD.Data + ":/var/lib/etcd "
	for key, value := range envs {
		cmd += fmt.Sprintf(" -e ETCD_%s=%s", strings.ToUpper(strings.ReplaceAll(key, "-", "_")), value)
	}
	for key, value := range ctlEnvs {
		cmd += fmt.Sprintf(" -e ETCDCTL_%s=%s", strings.ToUpper(strings.ReplaceAll(key, "-", "_")), value)
	}
	cmd += fmt.Sprintf(" %s etcd --name %s --data-dir /var/lib/etcd", image, node.Hostname)

	err := node.Sudo().Cmd(cmd)
	utils.Panic(err, "start etcd in docker")

	etcdPath := "/usr/local/bin/etcdctl"
	err = node.Sudo().ScpContent([]byte("#!/bin/bash\nset -e\n"+
		"docker exec -it vik8s-etcd /usr/local/bin/etcdctl $@"), etcdPath)
	utils.Panic(err, "make etcdctl command")

	err = node.Sudo().Cmd("chmod +x " + etcdPath)
	utils.Panic(err, "chmod Etcdctl command")
}

func restoreSnapshot(configure *config.Configuration, node *ssh.Node, image string) {
	if configure.ETCD.RemoteSnapshot != "" {
		logs.Infof("download etcd snapshot file: %s", configure.ETCD.RemoteSnapshot)

		resp, err := http.Get(configure.ETCD.RemoteSnapshot)
		utils.Panic(err, "etcd get remote snapshot")
		utils.Assert(resp.StatusCode == 200,
			"etcd get remote, the response status is %d not 200 %s", resp.StatusCode, resp.Status)
		defer resp.Body.Close()

		configure.ETCD.Snapshot = paths.Join("etcd", "snapshot.db")
		err = os.MkdirAll(filepath.Dir(configure.ETCD.Snapshot), os.ModePerm)
		utils.Panic(err, "make etcd config directory")

		fs, err := os.Create(configure.ETCD.Snapshot)
		utils.Panic(err, "etcd get remote snapshot")
		defer fs.Close()

		_, err = io.Copy(fs, resp.Body)
	}

	utils.Assert(configure.ETCD.Snapshot == "" || utils.Exists(configure.ETCD.Snapshot),
		"etcd snapshot file not found: %s", configure.ETCD.Snapshot)

	if configure.ETCD.Snapshot != "" {
		node.Logger("found etcd snapshot: %s", configure.ETCD.Snapshot)

		remotePath := node.HomeDir("snapshot.db")
		err := node.Scp(configure.ETCD.Snapshot, remotePath)
		utils.Panic(err, "upload etcd snapshot")

		restoreCmd := "docker run --rm --name etcd-restore-" + configure.ETCD.Token +
			" -v " + remotePath + ":/snapshot.db " +
			" -v " + filepath.Dir(configure.ETCD.Data) + ":/snapshot" +
			" " + image +
			" etcdctl snapshot restore /snapshot.db --data-dir /snapshot/etcd"
		err = node.Sudo().Cmd(restoreCmd)
		utils.Panic(err, "etcd load snapshot error")
	}
}

func initialCluster(configure *config.Configuration, node *ssh.Node) string {
	cluster := node.Hostname + "=https://" + node.Host + ":2380"
	for _, n := range hosts.MustGets(configure.ETCD.Nodes) {
		cluster += "," + n.Hostname + "=https://" + n.Host + ":2380"
	}
	return cluster
}

func waitEtcdReady(node *ssh.Node) {
	for i := 0; i < 5; i++ {
		status, _ := node.Sudo().CmdString("docker inspect vik8s-etcd -f '{{.State.Status}}'")
		if status == "running" {
			node.Logger("etcd node %s is ready", node.Host)
			return
		}
		node.Logger("etcd node %s status: %s", node.Host, status)
		time.Sleep(time.Second)
	}
	logs, err := node.Sudo().CmdString("docker logs --tail 10 vik8s-etcd")
	utils.Panic(utils.Wrap(err, logs), "")
}

func showClusterStatus(node *ssh.Node) {
	node.Logger("show etcd cluster")
	err := node.Sudo().CmdStdout(Etcdctl("endpoint status -w table"))
	utils.Panic(err, "show etcd cluster")

	err = node.Sudo().CmdStdout(Etcdctl("member list -w table"))
	utils.Panic(err, "show etcd cluster")
}

func Etcdctl(cmd string) string {
	return "docker exec vik8s-etcd /usr/local/bin/etcdctl " + cmd
}

func CertsDir() string {
	return paths.Join("etcd", "pki")
}
