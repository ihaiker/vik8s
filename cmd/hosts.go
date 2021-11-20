package cmd

import (
	"fmt"
	"github.com/ihaiker/cobrax"
	"github.com/ihaiker/vik8s/install/hosts"
	"github.com/ihaiker/vik8s/install/paths"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/spf13/cobra"
)

var _hosts_config = new(hosts.Option)

var hostsCmd = &cobra.Command{
	Use: "hosts", Short: "Add Management Host",
	Long: `vik8s hosts 172.16.100.4 172.16.100.10-172.16.100.15`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := hosts.Load(paths.HostsConfiguration(), _hosts_config)
		utils.Panic(err, "load host.conf error")

		_, err = hosts.Fetch(true, args...)
		utils.Panic(err, "add hosts")
	},
}

var hostsListCmd = &cobra.Command{
	Use: "list", Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		err := hosts.Load(paths.HostsConfiguration(), _hosts_config)
		utils.Panic(err, "load hosts.conf error")
		for _, node := range hosts.Nodes() {
			fmt.Println(node.Hostname, " ", node.Host)
		}
	},
}

func init() {
	err := cobrax.Flags(hostsCmd, _hosts_config, "", "VIK8S_SSH")
	utils.Panic(err, "set hosts command flag error")
	hostsCmd.AddCommand(hostsListCmd)
}
