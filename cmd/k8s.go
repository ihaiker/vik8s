package cmd

import (
	"fmt"
	"github.com/ihaiker/vik8s/cni"
	"github.com/ihaiker/vik8s/install/hosts"
	"github.com/ihaiker/vik8s/install/k8s"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	yamls "github.com/ihaiker/vik8s/yaml"
	"github.com/spf13/cobra"
	"time"
)

var initCmd = &cobra.Command{
	Use: "init", Short: "Initialize the kubernates cluster",
	Example: `vik8s init --master 172.10.0.2 --master 172.10.0.3 --master 172.10.0.4 --node 172.10.0.5 --ssh-user root --ssh-pk ~/.ssh/id_rsa
vik8s init -m 172.10.0.2 -m 172.10.0.3 -m 172.10.0.4 -n 172.10.0.5 -p password`,
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
	tenYear := time.Now().AddDate(44, 0, 0).Sub(time.Now())

	// Here you will define your flags and configuration settings.
	initCmd.Flags().IntVarP(&k8s.Config.SSH.Port, "ssh-port", "P", 22, "default port for ssh")
	initCmd.Flags().StringVarP(&k8s.Config.SSH.PrivateKey, "ssh-pk", "i", "$HOME/.ssh/id_rsa", "private key for ssh")
	initCmd.Flags().StringVarP(&k8s.Config.SSH.Password, "ssh-passwd", "p", "", "password for ssh\n")

	initCmd.Flags().StringSliceP("master", "m", []string{}, "k8s multi-masters. rule: XXX.XXX.XXX.XXX[-XXX.XXX.XXX.XXX][:PORT] ")
	initCmd.Flags().StringSliceP("node", "n", []string{}, "k8s multi-nodes. rule: XXX.XXX.XXX.XXX[-XXX.XXX.XXX.XXX][:PORT]\n")

	initCmd.Flags().StringVar(&k8s.Config.Docker.Version, "docker-version", "v1.19.12", "Specify docker version")
	initCmd.Flags().StringVar(&k8s.Config.Docker.DaemonJson, "docker-daemon-json", "", "docker config file it will overwirte `/etc/docker/daemon.json`")
	initCmd.Flags().StringSliceVar(&k8s.Config.Docker.Registry, "docker-registry-mirror", []string{}, "Customize docker registry, ignore it when set --docker-daemon-json")
	initCmd.Flags().BoolVar(&k8s.Config.Docker.StraitVersion, "docker-strait-version", false, "Strict check DOCKER version if inconsistent will upgrade\n")

	initCmd.Flags().StringVar(&k8s.Config.Kubernetes.KubeadmConfig, "kubeadm-config", "", "Path to a kubeadm configuration file. see kubeadm --config")
	initCmd.Flags().StringVar(&k8s.Config.Kubernetes.ApiServer, "apiserver", "vik8s-api-server", "Specify a stable IP address or DNS name for the control plane. see kubeadm  --control-plane-endpoint")
	initCmd.Flags().StringSliceVar(&k8s.Config.Kubernetes.ApiServerCertExtraSans, "apiserver-cert-extra-sans", []string{}, "see kubeadm init --apiserver-cert-extra-sans")
	initCmd.Flags().StringVar(&k8s.Config.Kubernetes.Version, "k8s-version", "v19.03.15", "Specify k8s version, support 1.17.+")
	initCmd.Flags().StringVar(&k8s.Config.Kubernetes.Interface, "interface", "eth.*|en.*|em.*", "name of network interface")
	initCmd.Flags().StringVar(&k8s.Config.Kubernetes.PodCIDR, "pod-cidr", "100.64.0.0/24", "Specify range of IP addresses for the pod network")
	initCmd.Flags().StringVar(&k8s.Config.Kubernetes.SvcCIDR, "svc-cidr", "10.96.0.0/12", "Use alternative range of IP address for service VIPs")
	initCmd.Flags().StringVar(&k8s.Config.Kubernetes.Repo, "repo", "", `Choose a container registry to pull control plane images from.
(default: Best choice from k8s.gcr.io and registry.aliyuncs.com/google_containers.)
`)

	initCmd.Flags().DurationVar(&k8s.Config.CertsValidity, "certs-validity", tenYear, "Certificate validity time")

	initCmd.Flags().BoolVar(&k8s.Config.ETCD.External, "etcd", false, `Use external ETCD cluster. 
If you installed the etcd cluster using 'vik8s etcd init', the cluster is used by default`)
	initCmd.Flags().StringSliceVar(&k8s.Config.ETCD.Nodes, "etcd-endpoints", []string{}, "the etcd cluster endpoints, for example: 172.16.100.10:2379")
	initCmd.Flags().StringVar(&k8s.Config.ETCD.CAFile, "etcd-ca", "", "the self-signed CA to provision identities for etcd")
	initCmd.Flags().StringVar(&k8s.Config.ETCD.ApiServerKeyFile, "etcd-apiserver-key-file", "", "the key file the apiserver uses to access etcd")
	initCmd.Flags().StringVar(&k8s.Config.ETCD.ApiServerCertFile, "etcd-apiserver-cert-file", "", "the certificate the apiserver uses to access etcd\n")

	initCmd.Flags().StringVar(&k8s.Config.Timezone, "timezone", "Asia/Shanghai", "")
	initCmd.Flags().StringSliceVar(&k8s.Config.NTPServices, "ntp-services", []string{"ntp1.aliyun.com", "ntp2.aliyun.com", "ntp3.aliyun.com"}, "time server")

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
