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
	Apply(node *ssh.Node)

	//清楚插件内容
	Clean(node *ssh.Node)
}

const DefaultPlugin = "flannel"

var Plugins = plugins{
	new(flannel), new(calico),
	new(customer),
}

func flags(f Plugin, name string) string {
	return fmt.Sprintf("cni-%s-%s", f.Name(), name)
}
