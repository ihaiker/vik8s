package etcd

import (
	"fmt"
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/install/hosts"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"gopkg.in/fatih/color.v1"
)

func ResetCluster(node *ssh.Node) {
	idx := utils.Search(config.Config.ETCD.Nodes, node.Host)
	if idx == -1 {
		fmt.Printf("%s not in the cluster\n", color.New(color.FgRed).Sprint(node.Host))
		return
	}

	removeEtcdMember(node)

	config.Config.ETCD.Nodes =
		append(config.Config.ETCD.Nodes[:idx], config.Config.ETCD.Nodes[idx+1:]...)

	if len(config.Config.ETCD.Nodes) != 0 {
		node := hosts.Get(config.Config.ETCD.Nodes[0])
		showClusterStatus(node)
	} else {
		utils.Line("all etcd node remove")
	}
}

func removeEtcdMember(node *ssh.Node) {
	node.Logger("remove etcd node %s", node.Host)

	id, err := node.SudoCmdString("docker exec vik8s-etcd " +
		"/usr/local/bin/Etcdctl member list | grep " + node.Host + ":2380 | awk -F',' '{print $1}'")
	utils.Panic(err, "etcd list member")

	if id != "" {
		if len(config.Config.ETCD.Nodes) != 1 {
			node.Logger("remove etcd member %s", id)
			err = node.SudoCmdPrefixStdout("docker exec vik8s-etcd /usr/local/bin/Etcdctl member remove " + id)
			utils.Panic(err, "etcd add member")
		}
	} else {
		node.Logger("this etcd node not found: %s", node.Host)
	}

	cleanEtcdData(node)
}

func cleanEtcdData(node *ssh.Node) {
	node.Logger("remove docker container vik8s-etcd")
	_ = node.SudoCmdPrefixStdout("docker rm -vf vik8s-etcd")

	node.Logger("remove etcd member data %s", config.Config.ETCD.Data)
	_ = node.SudoCmdPrefixStdout("rm -rf " + config.Config.ETCD.Data)
}
