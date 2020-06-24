package cmd

import (
	"fmt"
	"github.com/ihaiker/vik8s/install/hosts"
	"github.com/spf13/cobra"
	"os"
)

var hostsCmd = &cobra.Command{
	Use: "hosts", Short: "Add Management Host",
	Long: `vik8s hosts 172.16.100.4 172.16.100.10-172.16.100.15`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ssh := hosts.SSH{}
		ssh.Port, _ = cmd.Flags().GetInt("port")
		ssh.PkFile, _ = cmd.Flags().GetString("pk")
		ssh.Password, _ = cmd.Flags().GetString("passwd")
		hosts.Add(ssh, args...)
	},
}

var hostsListCmd = &cobra.Command{
	Use: "list", Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		for _, node := range hosts.Nodes() {
			fmt.Println(node.Hostname, " ", node.Host)
		}
	},
}

func init() {
	hostsCmd.Flags().Int("port", 22, "default port for ssh")
	hostsCmd.Flags().String("pk", os.ExpandEnv("$HOME/.ssh/id_rsa"), "private key for ssh")
	hostsCmd.Flags().String("passwd", "", "password for ssh")
	hostsCmd.AddCommand(hostsListCmd)
}
