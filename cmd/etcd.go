package cmd

import (
	"fmt"
	"github.com/ihaiker/cobrax"
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/install/etcd"
	"github.com/ihaiker/vik8s/install/hosts"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/spf13/cobra"
)

var etcdCmd = &cobra.Command{
	Use: "etcd", Short: "Install ETCD cluster",
	Long: `Install ETCD cluster.
This program uses etcdadm for installation, for details https://github.com/kubernetes-sigs/etcdadm`,
}

var etcdConfig = config.DefaultETCDConfiguration()
var etcdInitCmd = &cobra.Command{
	Use: "init", Short: "Initialize a new etcd cluster", Args: cobra.MinimumNArgs(1),
	PreRunE: configLoad(hostsLoad(none)), PostRunE: configDown(none),
	Example: `  vik8s etcd init 172.16.100.11-172.16.100.13
  vik8s etcd init 172.16.100.11 172.16.100.12 172.16.100.13`,
	Run: func(cmd *cobra.Command, args []string) {
		if configure.ETCD != nil && configure.ETCD.Token != "" {
			panic("The cluster has been initialized, if you want to re-initialize. use `vik8s etcd reset all` first ")
		}

		configure.ETCD = etcdConfig
		nodes := hosts.MustGets(args)
		etcd.InitCluster(configure, nodes[0])
		configure.ETCD.Nodes = append(configure.ETCD.Nodes, nodes[0].Host)

		args = make([]string, len(nodes)-1)
		for i, node := range nodes[1:] {
			args[i] = node.Host
		}
		etcdJoinCmd.Run(cmd, args)
	},
}

func init() {
	err := cobrax.Flags(etcdInitCmd, etcdConfig, "", "VIK8S_ETCD")
	utils.Panic(err, "set etcd configure")
	etcdInitCmd.Flags().SortFlags = false
}

var etcdJoinCmd = &cobra.Command{
	Use: "join", Short: "join nodes to etcd cluster",
	Example: `vik8s etcd join 172.16.100.10 172.16.100.11-172.16.100.13`,
	Args:    cobra.MinimumNArgs(1),
	PreRunE: configLoad(hostsLoad(none)), PostRunE: configDown(none),
	Run: func(cmd *cobra.Command, args []string) {
		utils.Assert(configure.ETCD != nil && len(configure.ETCD.Nodes) != 0,
			"etcd cluster not initialized yet")
		nodes := hosts.MustGets(args)
		for _, node := range nodes {
			utils.Assert(utils.Search(configure.ETCD.Nodes, node.Host) == -1,
				"has joined %s", node.Host)
			etcd.JoinCluster(configure, node)
			configure.ETCD.Nodes = append(configure.ETCD.Nodes, node.Host)
		}
		fmt.Println("-=-=-=- SUCCESS -=-=-=-")
	},
}

var etcdResetCmd = &cobra.Command{
	Use: "reset", Short: "reset etcd cluster",
	Example: `reset all: vik8s etcd reset
reset one node: vik8s etcd reset 172.16.100.10`,
	Args: cobra.MinimumNArgs(1), PreRunE: configLoad(hostsLoad(none)), PostRunE: configDown(none),
	Run: func(cmd *cobra.Command, args []string) {
		if configure.ETCD == nil {
			configure.ETCD = config.DefaultETCDConfiguration()
		}

		var nodes []*ssh.Node
		if len(args) == 1 && args[0] == "all" {
			nodes = hosts.MustGets(utils.Reverse(configure.ETCD.Nodes))
		} else {
			nodes = hosts.MustGets(args)
		}

		for _, node := range nodes {
			etcd.ResetCluster(configure, node)
		}

		if len(configure.ETCD.Nodes) == 0 {
			configure.ETCD = nil
		}
		fmt.Println("-=-=-=- SUCCESS -=-=-=-")
	},
}

func init() {
	etcdCmd.AddCommand(etcdInitCmd, etcdJoinCmd, etcdResetCmd)
}
