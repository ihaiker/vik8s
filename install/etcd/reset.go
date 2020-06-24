package etcd

import (
	"fmt"
	"github.com/ihaiker/vik8s/libs/ssh"
	"gopkg.in/fatih/color.v1"
)

func ResetCluster(node *ssh.Node) {
	if !Config.Exists(node.Host) {
		fmt.Printf("%s not in the cluster\n", color.New(color.FgRed).Sprint(node.Host))
		return
	}
	_ = node.MustCmd2String("etcdadm reset")
	Config.Remove(node.Host)
}
