package cni

import (
	"fmt"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/spf13/cobra"
)

type Plugin interface {
	Name() string

	//为初始化添加命令行参数
	Flags(cmd *cobra.Command)

	//生成插件
	Apply(cmd *cobra.Command, node *ssh.Node)

	//清楚插件内容
	Clean(node *ssh.Node)
}

type plugins []Plugin

var Plugins = plugins{
	NewFlannelCni(), new(calico),
	new(customer),
}

func (p *plugins) Apply(cmd *cobra.Command, node *ssh.Node) {
	for _, plugin := range *p {
		if plugin.Name() == cmd.Use {
			plugin.Apply(cmd, node)
		}
	}
}

func (p *plugins) Clean(node *ssh.Node) {
	_ = node.SudoCmd("ifconfig | grep cni0 > /dev/null && ifconfig cni0 down")
	_ = node.SudoCmd("ip link show | grep kube-ipvs0 && ip link delete kube-ipvs0 ")
	_ = node.SudoCmd("ip link show | grep dummy0 && ip link delete dummy0 ")
	for _, plugin := range *p {
		plugin.Clean(node)
	}
}

func flags(f Plugin, name string) string {
	return fmt.Sprintf("cni-%s-%s", f.Name(), name)
}
