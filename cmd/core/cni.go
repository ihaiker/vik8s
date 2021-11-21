package core

import (
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/install/cni"
	"github.com/ihaiker/vik8s/install/hosts"
	"github.com/spf13/cobra"
)

var cniCmd = &cobra.Command{
	Use: "cni", Short: "define kubernetes network interface",
}

func init() {
	for _, plugin := range cni.Plugins {
		name := plugin.Name()
		cmd := &cobra.Command{Use: name}
		plugin.Flags(cmd)
		cmd.PersistentPreRunE = configLoad(hostsLoad(none))
		cmd.PersistentPostRunE = configDown(none)
		cmd.Run = func(cmd *cobra.Command, args []string) {
			master := hosts.MustGet(config.K8S().Masters[0])
			cni.Plugins.Apply(cmd, master)
		}
		cniCmd.AddCommand(cmd)
	}
	rootCmd.AddCommand(cniCmd)
}
