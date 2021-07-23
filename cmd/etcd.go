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

func init() {
	etcdCmd.AddCommand(etcdInitCmd, etcdJoinCmd, etcdResetCmd)
}

var etcdConfig = new(config.ETCD)
var etcdInitCmd = &cobra.Command{
	Use: "init", Short: "Initialize a new etcd cluster",
	Args: cobra.MinimumNArgs(1), PersistentPreRunE: configLoad(hostsLoad(none)),
	Example: `  Args rule： [root:password@][ip-]ip[:port]
	vik8s etcd init 172.16.100.11-172.16.100.13
	vik8s etcd init 172.16.100.11 172.16.100.12 172.16.100.13`,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		ips := hosts.Add(args...)
		etcd.InitCluster(ips[0])
		for _, ip := range ips[1:] {
			etcd.JoinCluster(ip)
		}
		fmt.Println("-=-=-=- SUCCESS -=-=-=-")
		return
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
	Args:    cobra.MinimumNArgs(1), PersistentPreRunE: configLoad(hostsLoad(none)),
	Run: func(cmd *cobra.Command, args []string) {
		etcd.Config.MustRead()
		ips := hosts.Add(args...)
		for _, ip := range ips {
			etcd.JoinCluster(ip)
		}
		fmt.Println("-=-=-=- SUCCESS -=-=-=-")
	},
}

var etcdResetCmd = &cobra.Command{
	Use: "reset", Short: "reset etcd cluster",
	Example: `reset all: vik8s etcd reset
reset one node: vik8s etcd reset 172.16.100.10`, PersistentPreRunE: configLoad(hostsLoad(none)),
	Run: func(cmd *cobra.Command, args []string) {
		etcd.Config.MustRead()
		ips := utils.ParseIPS(args)
		if len(ips) == 0 {
			ips = etcd.Config.Nodes
		}
		nodes := hosts.Add(ips...)
		for _, node := range nodes {
			etcd.ResetCluster(node)
		}
		fmt.Println("-=-=-=- SUCCESS -=-=-=-")
	},
}
