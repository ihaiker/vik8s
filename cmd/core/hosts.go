package core

import (
	"fmt"
	"github.com/ihaiker/cobrax"
	hs "github.com/ihaiker/vik8s/install/hosts"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/spf13/cobra"
)

var _hosts_config = new(hs.Option)
var hostsCmd = &cobra.Command{
	Use: "hosts", Short: "Add Management Host",
	Long:    `vik8s hosts 172.16.100.4 172.16.100.10-172.16.100.15`,
	Args:    cobra.MinimumNArgs(1),
	PreRunE: configLoad(none),
	Run: func(cmd *cobra.Command, args []string) {
		_, err := configure.Hosts.Fetch(true, args...)
		utils.Panic(err, "add hosts")
	},
}

var hostsListCmd = &cobra.Command{
	Use: "list", Aliases: []string{"ls"},
	PreRunE: configLoad(none),
	Run: func(cmd *cobra.Command, args []string) {
		for _, node := range configure.Hosts.All() {
			fmt.Println(node.Hostname, " ", node.Host)
		}
	},
}

func init() {
	_ = cobrax.Flags(hostsCmd, _hosts_config, "", "VIK8S_SSH")
	//utils.Panic(err, "set hosts command flag error")
	hostsCmd.AddCommand(hostsListCmd)
}
