package etcd

import (
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/install/bases"
	"github.com/ihaiker/vik8s/install/cri"
	"github.com/ihaiker/vik8s/install/hosts"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
)

func JoinCluster(node *ssh.Node) {
	bases.Check(node)
	cri.Install(node)
	image := pullContainerImage(node)
	cleanEtcdData(node)
	makeAndPushCerts(node)
	addEtcdMember(node)
	joinEtcd(node, image)
	waitEtcdReady(node)
	showClusterStatus(node)
	config.Config.ETCD.Nodes = append(config.Config.ETCD.Nodes, node.Host)
}

func joinEtcd(node *ssh.Node, image string) {
	if config.Config.IsDockerCri() {
		initEtcdDocker(node, image, "existing")
	}
}

func addEtcdMember(node *ssh.Node) {
	node.Logger("add etcd node")
	master := hosts.Get(config.Config.ETCD.Nodes[0])
	num, err := master.SudoCmdString(etcdctl("member list | grep " + node.Host + ":2380 | wc -l"))
	utils.Panic(err, "etcd list member")
	if num == "0" {
		err = master.SudoCmdPrefixStdout(etcdctl("member add " + node.Hostname +
			" --peer-urls https://" + node.Host + ":2380"))
		utils.Panic(err, "etcd add member")
	}
}
