package etcd

import (
	"fmt"
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"gopkg.in/fatih/color.v1"
)

func ResetCluster(node *ssh.Node) {
	if utils.Search(config.Config.ETCD.Nodes, node.Host) == -1 {
		fmt.Printf("%s not in the cluster\n", color.New(color.FgRed).Sprint(node.Host))
		return
	}
}
