package etcd

import (
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
)

func ResetCluster(configure *config.Configuration, node *ssh.Node) {

	if utils.Search(configure.ETCD.Nodes, node.Host) != -1 {
		removeEtcdMember(configure, node)
		configure.ETCD.RemoveNode(node.Host)
	}

	cleanEtcdData(configure, node)

	if len(configure.ETCD.Nodes) > 0 {
		otherNode := configure.Hosts.MustGet(configure.ETCD.Nodes[0])
		showClusterStatus(otherNode)
	} else {
		utils.Line("all etcd node remove")
	}
}

func removeEtcdMember(configure *config.Configuration, node *ssh.Node) {
	if len(configure.ETCD.Nodes) == 1 {
		return
	}
	master := configure.Hosts.MustGet(utils.Any(configure.ETCD.Nodes, node.Host))

	id, err := master.Sudo().CmdString(Etcdctl("member list | grep " + node.Host + ":2380 | awk -F',' '{print $1}'"))
	utils.Panic(err, "etcd list member")

	if id != "" {
		node.Logger("remove etcd node %s", node.Host)
		if len(configure.ETCD.Nodes) != 1 {
			node.Logger("remove etcd member %s", id)
			err = master.Sudo().CmdPrefixStdout(Etcdctl("member remove " + id))
			utils.Panic(err, "etcd remove member")
		}
	} else {
		node.Logger("this etcd node not found: %s", node.Host)
	}
}

func cleanEtcdData(configure *config.Configuration, node *ssh.Node) {
	node.Logger("remove docker container vik8s-etcd")
	_ = node.Sudo().CmdPrefixStdout("docker rm -vf vik8s-etcd")

	node.Logger("remove etcd member data %s", configure.ETCD.DataRoot)
	_ = node.Sudo().CmdPrefixStdout("rm -rf " + configure.ETCD.DataRoot)

	node.Logger("remove etcd config data %s", configure.ETCD.CertsDir)
	_ = node.Sudo().CmdPrefixStdout("rm -rf " + configure.ETCD.CertsDir)
}
