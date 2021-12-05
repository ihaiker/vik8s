package core

import (
	"fmt"
	"github.com/ihaiker/cobrax"
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/install/etcd"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/spf13/cobra"
)

var etcdCmd = &cobra.Command{
	Use: "etcd", Short: "Install ETCD cluster",
	Long: `Install ETCD cluster`,
}

var etcdConfig = config.DefaultETCDConfiguration()
var etcdInitCmd = &cobra.Command{
	Use: "init", Short: "Initialize a new etcd cluster", Args: cobra.MinimumNArgs(1),
	PreRunE: configLoad(none), PostRunE: configDown(none),
	Example: `  vik8s etcd init 172.16.100.11-172.16.100.13
  vik8s etcd init 172.16.100.11 172.16.100.12 172.16.100.13`,
	Run: func(cmd *cobra.Command, args []string) {
		if configure.ETCD != nil && configure.ETCD.Token != "" {
			panic("The cluster has been initialized, if you want to re-initialize. use `vik8s etcd reset all` first ")
		}

		configure.ETCD = etcdConfig
		nodes := configure.Hosts.MustGets(args)
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
	PreRunE: configLoad(none), PostRunE: configDown(none),
	Run: func(cmd *cobra.Command, args []string) {
		utils.Assert(configure.ETCD != nil && len(configure.ETCD.Nodes) != 0,
			"etcd cluster not initialized yet")
		nodes := configure.Hosts.MustGets(args)
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
	Args: cobra.MinimumNArgs(1), PreRunE: configLoad(none), PostRunE: configDown(none),
	Run: func(cmd *cobra.Command, args []string) {
		if configure.ETCD == nil {
			configure.ETCD = config.DefaultETCDConfiguration()
		}

		var nodes []*ssh.Node
		if len(args) == 1 && args[0] == "all" {
			nodes = configure.Hosts.MustGets(utils.Reverse(configure.ETCD.Nodes))
		} else {
			nodes = configure.Hosts.MustGets(args)
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
var externalETCDConfigure = new(config.ExternalETCDConfiguration)
var externalETCDCmd = &cobra.Command{
	Use: "external", Short: "use external etcd cluster",
	Example: `vikis etcd external --endpoint https://127.0.0.1:2182`,
	PreRunE: configLoad(none), PostRunE: configDown(none),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(externalETCDConfigure.Endpoints) == 0 {
			return utils.Error("At least one endpoint is required")
		}
		if externalETCDConfigure.Cert == "" {
			return utils.Error("--cert is required")
		}
		if utils.NotExists(externalETCDConfigure.Cert) {
			return utils.Error("cert %s not found", externalETCDConfigure.Cert)
		}
		if externalETCDConfigure.Key == "" {
			return utils.Error("--key is required")
		}
		if utils.NotExists(externalETCDConfigure.Key) {
			return utils.Error("key %s not found", externalETCDConfigure.Key)
		}
		if externalETCDConfigure.CaFile == "" {
			return utils.Error("--ca is required")
		}
		if utils.NotExists(externalETCDConfigure.CaFile) {
			return utils.Error("ca %s not found", externalETCDConfigure.CaFile)
		}
		configure.ExternalETCD = externalETCDConfigure
		return nil
	},
}

func init() {
	err := cobrax.Flags(externalETCDCmd, externalETCDConfigure, "", "VIK8S_EXTERNAL_ETCD")
	utils.Panic(err, "set etcd configure")
	etcdCmd.AddCommand(etcdInitCmd, etcdJoinCmd, etcdResetCmd, externalETCDCmd)
}
