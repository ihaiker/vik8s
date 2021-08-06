package cmd

import (
	"fmt"
	"github.com/ihaiker/cobrax"
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/install/cni"
	"github.com/ihaiker/vik8s/install/hosts"
	"github.com/ihaiker/vik8s/install/k8s"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	yamls "github.com/ihaiker/vik8s/yaml"
	"github.com/spf13/cobra"
)

var k8sConfig = config.DefaultK8SConfiguration()
var initCmd = &cobra.Command{
	Use: "init", Short: "Initialize the kubernates cluster",
	Example: `vik8s init --master 172.10.0.2 --master 172.10.0.3 --master 172.10.0.4 --node 172.10.0.5 --ssh-user root --ssh-pk ~/.ssh/id_rsa
vik8s init -m 172.10.0.2 -m 172.10.0.3 -m 172.10.0.4 -n 172.10.0.5 -p password`,
	PreRunE: configLoad(hostsLoad(none)), PostRunE: configDown(none),
	Run: func(cmd *cobra.Command, args []string) {
		k8s.Config.Parse()
		masters := getFlagsIps(cmd, "master")
		nodes := getFlagsIps(cmd, "node")

		utils.Assert(len(masters) != 0, "master node is empty")
		master := k8s.InitCluster(masters[0])
		cni.Plugins.Apply(master)

		for _, ctl := range masters[1:] {
			k8s.JoinControl(ctl)
		}
		for _, node := range nodes {
			k8s.JoinWorker(node)
		}
		fmt.Println("-=-=-=- SUCCESS -=-=-=-")
	},
}

func init() {
	err := cobrax.FlagsWith(initCmd, cobrax.GetFlags, k8sConfig, "", "VIK8S_K8S")
	utils.Panic(err, "setting `init` flag error")
	cni.Plugins.Flags(initCmd)
	initCmd.Flags().SortFlags = false
}

var joinCmd = &cobra.Command{
	Use: "join", Short: "join to k8s",
	Example: `vik8s join --master 172.10.0.2-172.10.0.4 --node 172.10.0.7
vik8s join -m 172.10.0.2 -m 172.10.0.3 -m 172.10.0.4 -n 172.10.0.5`,
	PreRun: k8s.Config.LoadCmd,
	Run: func(cmd *cobra.Command, args []string) {
		masters := getFlagsIps(cmd, "master")
		nodes := getFlagsIps(cmd, "node")
		isAsync, _ := cmd.Flags().GetBool("async")

		if len(masters) == 0 && len(nodes) == 0 {
			fmt.Println(cmd.UseLine())
			return
		}

		async := utils.Async()
		for _, ctl := range masters {
			if isAsync {
				async.Add(k8s.JoinControl, ctl)
			} else {
				k8s.JoinControl(ctl)
			}
		}

		for _, node := range nodes {
			if isAsync {
				async.Add(k8s.JoinWorker, node)
			} else {
				k8s.JoinWorker(node)
			}
		}
		async.Wait()
		fmt.Println("-=-=-=- SUCCESS -=-=-=-")
	},
}

func init() {
	joinCmd.Flags().StringSliceP("master", "m", []string{}, "")
	joinCmd.Flags().StringSliceP("node", "n", []string{}, "")
	joinCmd.Flags().Bool("async", false, "Whether to execute asynchronously")
}

var resetCmd = &cobra.Command{
	Use: "reset", Short: "reset",
	PreRun: k8s.Config.LoadCmd, Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		nodes := args
		if args[0] == "all" {
			nodes = append(k8s.Config.Nodes, utils.Reverse(k8s.Config.Masters)...)
		}
		for _, nodeName := range nodes {
			node := hosts.Get(nodeName)
			_, _ = k8s.Config.Master().
				Cmd(fmt.Sprintf("kubectl delete nodes %s", node.Hostname))
			k8s.ResetNode(node)
		}
		fmt.Println("-=-=-=- SUCCESS -=-=-=-")
	},
}

func init() {
	resetCmd.Flags().Bool("force", false, "")
}

var configCmd = &cobra.Command{
	Use: "config", Short: "Show yaml file used by vik8s deployment cluster",
	Args: cobra.ExactValidArgs(1), ValidArgs: yamls.AssetNames(),
	Example: "vik8s config all",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(string(yamls.MustAsset(args[0])))
	},
}
var configNamesCmd = &cobra.Command{
	Use: "names", Short: "show file names",
	Run: func(cmd *cobra.Command, args []string) {
		for _, name := range yamls.AssetNames() {
			fmt.Println(name)
		}
	},
}

func init() {
	configCmd.AddCommand(configNamesCmd)
}

func getFlagsIps(cmd *cobra.Command, name string) ssh.Nodes {
	values, err := cmd.Flags().GetStringSlice(name)
	utils.Panic(err, "get flags %s", name)
	return hosts.Add(values...)
}
