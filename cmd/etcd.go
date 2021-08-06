package cmd

import (
	"fmt"
	"github.com/ihaiker/cobrax"
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/install/etcd"
	"github.com/ihaiker/vik8s/install/hosts"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/spf13/cobra"
)

var etcdCmd = &cobra.Command{
	Use: "etcd", Short: "Install ETCD cluster",
	Long: `Install ETCD cluster.
This program uses etcdadm for installation, for details https://github.com/kubernetes-sigs/etcdadm`,
}

var etcdConfig = new(config.ETCD)
var etcdInitCmd = &cobra.Command{
	Use: "init", Short: "Initialize a new etcd cluster", Args: cobra.MinimumNArgs(1),
	PreRunE: configLoad(hostsLoad(none)), PostRunE: configDown(none),
	Example: `  vik8s etcd init 172.16.100.11-172.16.100.13
  vik8s etcd init 172.16.100.11 172.16.100.12 172.16.100.13`,
	Run: func(cmd *cobra.Command, args []string) {
		config.Config.ETCD = etcdConfig
		nodes := hosts.Add(args...)
		hosts.MustGatheringFacts(nodes...)
		etcd.InitCluster(nodes[0])
		for _, ip := range nodes[1:] {
			etcd.JoinCluster(ip)
		}
		fmt.Println("-=-=-=- SUCCESS -=-=-=-")
	},
}

func init() {
	err := cobrax.Flags(etcdInitCmd, etcdConfig, "", "VIK8S_ETCD")
	utils.Panic(err, "set etcd configuration")
	etcdInitCmd.Flags().SortFlags = false
}

var etcdJoinCmd = &cobra.Command{
	Use: "join", Short: "join nodes to etcd cluster",
	Example: `vik8s etcd join 172.16.100.10 172.16.100.11-172.16.100.13`,
	Args:    cobra.MinimumNArgs(1),
	PreRunE: configLoad(hostsLoad(none)), PostRunE: configDown(none),
	Run: func(cmd *cobra.Command, args []string) {
		utils.Assert(config.Config.ETCD != nil && len(config.Config.ETCD.Nodes) != 0,
			"etcd cluster not initialized yet")

		nodes := hosts.Add(args...)
		hosts.MustGatheringFacts(nodes...)
		for _, node := range nodes {
			utils.Assert(utils.Search(config.Config.ETCD.Nodes, node.Host) == -1,
				"has joined %s", node.Host)
			etcd.JoinCluster(node)
		}
		fmt.Println("-=-=-=- SUCCESS -=-=-=-")
	},
}

var etcdResetCmd = &cobra.Command{
	Use: "reset", Short: "reset etcd cluster",
	Example: `reset all: vik8s etcd reset
reset one node: vik8s etcd reset 172.16.100.10`,
	PreRunE: configLoad(hostsLoad(none)), PostRunE: configDown(none),
	Run: func(cmd *cobra.Command, args []string) {
		utils.Assert(config.Config.ETCD != nil && len(config.Config.ETCD.Nodes) > 0,
			"not found etcd cluster")

		nodes := hosts.Add(args...)
		if len(nodes) == 0 {
			nodes = hosts.Gets(config.Config.ETCD.Nodes)
		}
		for _, node := range nodes {
			etcd.ResetCluster(node)
		}
		if len(config.Config.ETCD.Nodes) == 0 {
			config.Config.ETCD = nil
		}
		fmt.Println("-=-=-=- SUCCESS -=-=-=-")
	},
}

func init() {
	etcdCmd.AddCommand(etcdInitCmd, etcdJoinCmd, etcdResetCmd)
}
