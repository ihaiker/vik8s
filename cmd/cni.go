package cmd

import (
	"github.com/ihaiker/vik8s/install/cni"
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
		cmd.PersistentPreRunE = configLoad(none)
		cmd.PersistentPostRunE = configDown(none)
		cmd.Run = func(cmd *cobra.Command, args []string) {
			master := configure.Hosts.MustGet(configure.K8S.Masters[0])
			cni.Plugins.Apply(cmd, configure, master)
		}
		cniCmd.AddCommand(cmd)
	}
	rootCmd.AddCommand(cniCmd)
}
