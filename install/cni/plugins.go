package cni

import (
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/spf13/cobra"
)

type Plugin interface {
	Name() string

	//Flags 为初始化添加命令行参数
	Flags(cmd *cobra.Command)

	//Apply 生成插件
	Apply(cmd *cobra.Command, configure *config.Configuration, node *ssh.Node)

	//Clean 清楚插件内容
	Clean(node *ssh.Node)
}

type plugins []Plugin

var Plugins = plugins{
	NewFlannelCni(), NewCalico(),
	new(customer),
}

func (p *plugins) Apply(cmd *cobra.Command, configure *config.Configuration, node *ssh.Node) {
	for _, plugin := range *p {
		if plugin.Name() == cmd.Use {
			plugin.Apply(cmd, configure, node)
		}
	}
}

func (p *plugins) Clean(node *ssh.Node) {
	_ = node.Sudo().Cmd("ifconfig | grep cni0 > /dev/null && ifconfig cni0 down")
	_ = node.Sudo().Cmd("ip link show | grep kube-ipvs0 && ip link delete kube-ipvs0 ")
	_ = node.Sudo().Cmd("ip link show | grep dummy0 && ip link delete dummy0 ")
	for _, plugin := range *p {
		plugin.Clean(node)
	}
}
