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
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func InitCluster(node *ssh.Node) {
	node.Logger("install etcd server")
	if config.Config.ETCD.Token == "" {
		token := utils.Random(16)
		node.Logger("make etcd token: %s", token)
		config.Config.ETCD.Token = token
	}
	bases.Check(node)
	cri.Install(node)
	image := pullContainerImage(node)
	makeAndPushCerts(node)
	initEtcd(node, image)
	waitEtcdReady(node)
	showClusterStatus(node)

	config.Config.ETCD.Nodes = append(config.Config.ETCD.Nodes, node.Host)
}

func pullContainerImage(node *ssh.Node) (image string) {
	if config.Config.IsDockerCri() {
		repoUrl := config.Config.ETCD.Repo
		image = fmt.Sprintf("%s/%s:%s", repo.QuayIO(repoUrl), "coreos/etcd", config.Config.ETCD.Version)
		num, err := node.SudoCmdString(fmt.Sprintf("docker images --format '{{.Repository}}:{{.Tag}}' | grep %s | wc -l", image))
		utils.Panic(err, "check docker image tag")
		if num == "0" {
			err = node.SudoCmdOutput("docker pull "+image, os.Stdout)
			utils.Panic(err, "pull docker image")
		}
	}
	return
}

func makeAndPushCerts(node *ssh.Node) {
	node.Logger("make certs files")

	name := node.Hostname
	dir := paths.Join("etcd", "pki")
	sans := []string{"127.0.0.1", "localhost", node.Hostname, node.Host}
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

func initEtcd(node *ssh.Node, image string) {
	if config.Config.IsDockerCri() {
		initEtcdDocker(node, image, "new")
	}
}

func initEtcdDocker(node *ssh.Node, image string, state string) {
	_ = node.SudoCmd("docker rm -vf vik8s-etcd")

	envs := map[string]string{
		"initial-advertise-peer-urls": "https://" + node.Host + ":2380",                        //对外通告该节点的同伴（Peer）监听地址，这个值会告诉集群中其他节点。
		"listen-peer-urls":            "https://" + node.Host + ":2380",                        //指定和 Cluster 其他 Node 通信的地址
		"listen-client-urls":          "https://" + node.Host + ":2379,https://127.0.0.1:2379", //指定对外提供服务的地址
		"advertise-client-urls":       "https://" + node.Host + ":2380",                        //对外通告的该节点的客户端监听地址，会告诉集群中其他节点。

		"initial-cluster-token": config.Config.ETCD.Token, //创建集群
		"initial-cluster-state": state,                    //初始化新集群时使用 new, 加入已有集群时使用 existing
		"initial-cluster":       initialCluster(node),     //指定集群成员列表

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
	cmd := "docker run -d --name vik8s-etcd --restart always --network host --hostname " + node.Hostname +
		" -v " + config.Config.ETCD.CertsDir + ":/etc/etcd/pki" +
		" -v " + config.Config.ETCD.Data + ":/var/lib/etcd "
	for key, value := range envs {
		cmd += fmt.Sprintf(" -e ETCD_%s=%s", strings.ToUpper(strings.ReplaceAll(key, "-", "_")), value)
	}
	for key, value := range ctlEnvs {
		cmd += fmt.Sprintf(" -e ETCDCTL_%s=%s", strings.ToUpper(strings.ReplaceAll(key, "-", "_")), value)
	}
	cmd += fmt.Sprintf(" %s etcd --name %s --data-dir /var/lib/etcd", image, node.Hostname)

	err := node.SudoCmd(cmd)
	utils.Panic(err, "start etcd in docker")

	etcdPath := "/usr/local/bin/etcdctl"
	err = node.SudoScpContent([]byte("#!/bin/bash\nset -e\n"+
		"docker exec -it vik8s-etcd /usr/local/bin/etcdctl $@"), etcdPath)
	utils.Panic(err, "make etcdctl command")

	err = node.SudoCmd("chmod +x " + etcdPath)
	utils.Panic(err, "chmod etcdctl command")
}

func initialCluster(node *ssh.Node) string {
	cluster := node.Hostname + "=https://" + node.Host + ":2380"
	for _, n := range hosts.Gets(config.Config.ETCD.Nodes) {
		cluster += "," + n.Hostname + "=https://" + n.Host + ":2380"
	}
	return cluster
}

func waitEtcdReady(node *ssh.Node) {
	for i := 0; i < 5; i++ {
		status, _ := node.SudoCmdString("docker inspect vik8s-etcd -f '{{.State.Status}}'")
		if status == "running" {
			node.Logger("etcd node %s is ready", node.Host)
			return
		}
		node.Logger("etcd node %s status: %s", node.Host, status)
		time.Sleep(time.Second)
	}
	logs, err := node.SudoCmdString("docker logs --tail 10 vik8s-etcd")
	utils.Panic(utils.Wrap(err, logs), "")
}

func showClusterStatus(node *ssh.Node) {
	node.Logger("show etcd cluster")
	err := node.SudoCmdStdout("docker exec vik8s-etcd /usr/local/bin/etcdctl endpoint status -w table")
	utils.Panic(err, "show etcd cluster")

	err = node.SudoCmdStdout("docker exec vik8s-etcd /usr/local/bin/etcdctl member list -w table")
	utils.Panic(err, "show etcd cluster")
}
