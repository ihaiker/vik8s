package cmd

import (
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/ingress"
	"github.com/ihaiker/vik8s/install/hosts"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/spf13/cobra"
)

var ingressRootCmd = &cobra.Command{
	Use: "ingress", Short: "install kubernetes ingress controller",
	Example: "vik8s ingress nginx",
	PreRunE: configLoad(hostsLoad(none)), PostRunE: configDown(none),
}

func uninstallIngressCmd() *cobra.Command {
	return &cobra.Command{
		Use: "uninstall", Aliases: []string{"remove", "delete", "del"},
		Run: func(cmd *cobra.Command, args []string) {
			master := hosts.Get(config.K8S().Masters[0])
			name := cmd.Parent().Name()
			ingress.Manager.Delete(name, master)
		},
	}
}

func ingressRun(cmd *cobra.Command, args []string) {
	master := hosts.Get(config.K8S().Masters[0])
	name := cmd.Name()
	ingress.Manager.Apply(name, master)
}

func init() {
	for _, plugin := range ingress.Manager {
		//install
		cmd := &cobra.Command{
			Use: plugin.Name(), Short: utils.FirstLine(plugin.Description()),
			Long: plugin.Description(), Run: ingressRun,
		}
		plugin.Flags(cmd)
		cmd.Flags().SortFlags = false
		ingressRootCmd.AddCommand(cmd)
		//uninstall
		cmd.AddCommand(uninstallIngressCmd())
	}
}
