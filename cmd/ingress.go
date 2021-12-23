package cmd

import (
	"github.com/ihaiker/vik8s/ingress"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/spf13/cobra"
)

var ingressRootCmd = &cobra.Command{
	Use: "ingress", Short: "install kubernetes ingress controller",
	Example: "vik8s ingress nginx",
}

func ingressRun(cmd *cobra.Command, args []string) {
	master := configure.Hosts.MustGet(configure.K8S.Masters[0])
	name := cmd.Name()
	ingress.Manager.Apply(name, master)
}

func init() {
	for _, plugin := range ingress.Manager {
		//install
		cmd := &cobra.Command{
			Use: plugin.Name(), Short: utils.FirstLine(plugin.Description()),
			Long: plugin.Description(), Run: ingressRun,
			PreRunE: configLoad(none), PostRunE: configDown(none),
		}
		plugin.Flags(cmd)
		cmd.Flags().SortFlags = false
		ingressRootCmd.AddCommand(cmd)
	}
}
