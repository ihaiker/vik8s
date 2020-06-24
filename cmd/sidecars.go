package cmd

import (
	"github.com/ihaiker/vik8s/install/k8s"
	"github.com/ihaiker/vik8s/sidecars"
	"github.com/spf13/cobra"
	"strings"
)

var sidecarsCmd = &cobra.Command{
	Use: "sidecars", Aliases: []string{"ss"},
	PersistentPreRun: k8s.Config.LoadCmd,
}

func uninstallSidecarsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "uninstall", Run: func(cmd *cobra.Command, args []string) {
			name := cmd.Parent().Name()
			data, _ := cmd.Flags().GetBool("data")
			sidecars.Manager.Delete(name, data)
		},
	}
	cmd.Flags().Bool("data", false, "remove data folder")
	return cmd
}

func init() {
	for _, plugin := range sidecars.Manager {
		cmd := &cobra.Command{
			Use: plugin.Name(), Long: plugin.Description(),
			Short: strings.SplitN(plugin.Description(), "\n", 2)[0],
			Run: func(cmd *cobra.Command, args []string) {
				name := cmd.Name()
				sidecars.Manager.Apply(name)
			},
		}
		plugin.Flags(cmd)
		cmd.Flags().SortFlags = false
		sidecarsCmd.AddCommand(cmd)
		cmd.AddCommand(uninstallSidecarsCmd())
	}
}
