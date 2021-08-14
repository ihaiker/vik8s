package etcd

import (
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/install/hosts"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
)

func ResetCluster(node *ssh.Node) {

	idx := utils.Search(config.Config.ETCD.Nodes, node.Host)
	if idx != -1 {
		removeEtcdMember(node)
		config.Config.ETCD.Nodes =
			append(config.Config.ETCD.Nodes[:idx], config.Config.ETCD.Nodes[idx+1:]...)
	}

	cleanEtcdData(node)

	if len(config.Config.ETCD.Nodes) > 0 {
		otherNode := hosts.Get(config.Config.ETCD.Nodes[0])
		showClusterStatus(otherNode)
	} else {
		utils.Line("all etcd node remove")
	}
}

func removeEtcdMember(node *ssh.Node) {
	node.Logger("remove etcd node %s", node.Host)

	id, err := node.Sudo().CmdString(Etcdctl("member list | grep " + node.Host + ":2380 | awk -F',' '{print $1}'"))
	utils.Panic(err, "etcd list member")

	if id != "" {
		if len(config.Config.ETCD.Nodes) != 1 {
			node.Logger("remove etcd member %s", id)
			err = node.Sudo().CmdPrefixStdout(Etcdctl("member remove " + id))
			utils.Panic(err, "etcd add member")
		}
	} else {
		node.Logger("this etcd node not found: %s", node.Host)
	}
}

func cleanEtcdData(node *ssh.Node) {
	node.Logger("remove docker container vik8s-etcd")
	_ = node.Sudo().CmdPrefixStdout("docker rm -vf vik8s-etcd")

	node.Logger("remove etcd member data %s", config.Config.ETCD.Data)
	_ = node.Sudo().CmdPrefixStdout("rm -rf " + config.Config.ETCD.Data)

	node.Logger("remove etcd config data %s", config.Config.ETCD.CertsDir)
	_ = node.Sudo().CmdPrefixStdout("rm -rf " + config.Config.ETCD.CertsDir)
}
