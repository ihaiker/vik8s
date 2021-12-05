package etcd

import (
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/install/bases"
	"github.com/ihaiker/vik8s/install/cri"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
)

func JoinCluster(configure *config.Configuration, node *ssh.Node) {
	bases.Check(node)
	cri.Install(configure, node)
	image := pullContainerImage(configure, node)

	removeEtcdMember(configure, node)
	cleanEtcdData(configure, node)

	makeAndPushCerts(configure, node)
	addEtcdMember(configure, node)
	joinEtcd(configure, node, image)
	waitEtcdReady(node)
	showClusterStatus(node)
}

func joinEtcd(configure *config.Configuration, node *ssh.Node, image string) {
	if configure.IsDockerCri() {
		initEtcdDocker(configure, node, image, "existing")
	}
}

func addEtcdMember(configure *config.Configuration, node *ssh.Node) {
	node.Logger("add etcd node")
	master := configure.Hosts.MustGet(configure.ETCD.Nodes[0])
	num, err := master.Sudo().CmdString(Etcdctl("member list | grep " + node.Host + ":2380 | wc -l"))
	utils.Panic(err, "etcd list member")
	if num == "0" {
		err = master.Sudo().CmdPrefixStdout(Etcdctl("member add " + node.Hostname +
			" --peer-urls https://" + node.Host + ":2380"))
		utils.Panic(err, "etcd add member")
	}
}
