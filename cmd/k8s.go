package cmd

import (
	"fmt"
	"github.com/ihaiker/cobrax"
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/install/k8s"
	"github.com/ihaiker/vik8s/libs/logs"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/spf13/cobra"
	"strings"
)

var k8sConfig = config.DefaultK8SConfiguration()
var initCmd = &cobra.Command{
	Use: "init", Short: "Initialize the kubernetes cluster",
	Example: `vik8s init 10.24.1.10 172.10.0.2-172.10.0.5`,
	PreRunE: configLoad(none), PostRunE: configDown(none),
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		masters := configure.Hosts.MustGets(args)

		utils.Assert(len(masters) != 0, "master node is empty")
		configure.K8S = k8sConfig

		k8s.InitCluster(configure, masters[0])
		for _, ctl := range masters[1:] {
			k8s.JoinControl(configure, ctl)
		}

		taint, _ := cmd.Flags().GetBool("taint")
		if taint {
			for _, master := range masters {
				k8s.RemoveTaint(configure, master)
			}
		}
		fmt.Println("-=-=-=- SUCCESS -=-=-=-")
	},
}

func init() {
	err := cobrax.FlagsWith(initCmd, cobrax.GetFlags, k8sConfig, "", "VIK8S_K8S")
	utils.Panic(err, "setting `init` flag error")
	initCmd.Flags().Bool("taint", false, "Update the taints on the nodes")
	initCmd.Flags().SortFlags = false
}

var joinCmd = &cobra.Command{
	Use: "join", Short: "join to k8s",
	Example: `vik8s join --master 172.10.0.2-172.10.0.4
vik8s join 172.10.0.2 172.10.0.3 172.10.0.4 172.10.0.5`,
	PreRunE: configLoad(none), PostRunE: configDown(none),
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		nodes := configure.Hosts.MustGets(args)
		if len(nodes) == 0 {
			fmt.Println(cmd.UseLine())
			return
		}
		master, _ := cmd.Flags().GetBool("master")
		for _, node := range nodes {
			utils.Assert(utils.Search(configure.K8S.Masters, node.Host) == -1 &&
				utils.Search(configure.K8S.Nodes, node.Host) == -1,
				"the host is cluster node yet. %s", node.Hostname)

			if master {
				k8s.JoinControl(configure, node)
			} else {
				k8s.JoinWorker(configure, node)
			}
			taint, _ := cmd.Flags().GetBool("taint")
			if taint {
				k8s.RemoveTaint(configure, node)
			}
		}
		fmt.Println("-=-=-=- SUCCESS -=-=-=-")
	},
}

func init() {
	joinCmd.Flags().BoolP("master", "m", false, "Whether it is a control plane")
	joinCmd.Flags().Bool("taint", false, "Update the taints on the nodes")
}

var resetCmd = &cobra.Command{
	Use: "reset", Short: "reset kubernetes cluster node",
	Example: "vik8s reset <nodeIP|nodeHostName|all>",
	Args:    cobra.MinimumNArgs(1),
	PreRunE: configLoad(none), PostRunE: configDown(none),
	Run: func(cmd *cobra.Command, args []string) {
		nodes := args
		if configure.K8S == nil {
			configure.K8S = config.DefaultK8SConfiguration()
		}
		if args[0] == "all" {
			nodes = append(configure.K8S.Nodes, utils.Reverse(configure.K8S.Masters)...)
		}
		var master *ssh.Node
		if len(configure.K8S.Masters) > 0 {
			master = configure.Hosts.MustGet(configure.K8S.Masters[0])
		}
		for _, nodeName := range nodes {
			node := configure.Hosts.MustGet(nodeName)
			utils.Assert(node != nil, "not found kubernetes %s", node.Host)
			logs.Infof("remove cluster node %s", node.Prefix())

			if master != nil {
				err := master.Cmd(fmt.Sprintf("kubectl delete nodes %s", node.Hostname))
				utils.Assert(err == nil || strings.Contains(err.Error(), "not found"),
					"reset kubernetes node: %v", err)
			}
			k8s.ResetNode(configure, node)
		}
		fmt.Println("-=-=-=- SUCCESS -=-=-=-")
	},
}

func init() {
	resetCmd.Flags().Bool("force", false, "")
}
