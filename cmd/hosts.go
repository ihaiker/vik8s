package cmd

import (
	"fmt"
	"github.com/ihaiker/cobrax"
	"github.com/ihaiker/vik8s/install/hosts"
	"github.com/ihaiker/vik8s/install/paths"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/spf13/cobra"
	"os/user"
)

var _hosts_config = new(hosts.Option)

var hostsCmd = &cobra.Command{
	Use: "hosts", Short: "Add Management Host",
	Long: `vik8s hosts 172.16.100.4 172.16.100.10-172.16.100.15`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		manager, err := hosts.New(paths.HostsConfiguration(), _hosts_config, true)
		utils.Panic(err, "load hosts.conf error")
		_, err = manager.Add(args...)
		utils.Panic(err, "add host")
	},
}

var hostsListCmd = &cobra.Command{
	Use: "list", Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		manager, err := hosts.New(paths.HostsConfiguration(), _hosts_config, true)
		utils.Panic(err, "load hosts.conf error")
		for _, node := range manager.All() {
			fmt.Println(node.Hostname, " ", node.Host)
		}
	},
}

func init() {
	if user, err := user.Current(); err == nil {
		_hosts_config.User = user.Username
	} else {
		_hosts_config.User = "root"
	}
	err := cobrax.Flags(hostsCmd, _hosts_config, "", "VIK8S_SSH")
	utils.Panic(err, "set hosts command flag error")
	hostsCmd.AddCommand(hostsListCmd)
}
