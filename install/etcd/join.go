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
	master := hosts.Get(config.Config.ETCD.Nodes[0])

	bases.Check(node)
	cri.Install(node)
	pullContainerImage(node)
	makeAndPushCerts(node)

	etcdadmJoin(master, node)
}

func etcdadmJoin(master *ssh.Node, node *ssh.Node) {
	utils.Line("etcd join")
}
