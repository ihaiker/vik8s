package cni

import (
	"fmt"
	"github.com/ihaiker/vik8s/install/k8s"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/spf13/cobra"
	"os"
)

type plugins []Plugin

func (p *plugins) Name() string {
	return "plugins-manager"
}

func (p *plugins) Flags(cmd *cobra.Command) {
	support := "ignore"
	for _, plugin := range *p {
		support += "," + plugin.Name()
	}
	cmd.Flags().StringVar(&k8s.Config.CNI.Name, "cni", DefaultPlugin,
		fmt.Sprintf("the kubernates cni plugins. support: %s", support))
	for _, plugin := range *p {
		support += "," + plugin.Name()
		plugin.Flags(cmd)
	}
}

func (p *plugins) Apply(node *ssh.Node) {
	cni := k8s.Config.CNI.Name
	if cni == "ignore" || cni == "" {
		fmt.Println("[warn] cni plugin is ignore")
		return
	}
	for _, plugin := range *p {
		if cni == plugin.Name() {
			utils.Line("apply %s network", plugin.Name())
			plugin.Apply(node)
			return
		}
	}
	utils.Panic(os.ErrNotExist, "not found cni %s", cni)
}

func (p *plugins) Clean(node *ssh.Node) {
	_, _ = node.Cmd("ifconfig | grep cni0 > /dev/null && ifconfig cni0 down")
	_, _ = node.Cmd("ip link show | grep kube-ipvs0 && ip link delete kube-ipvs0 ")
	_, _ = node.Cmd("ip link show | grep dummy0 && ip link delete dummy0 ")
	for _, plugin := range *p {
		plugin.Clean(node)
	}
}
