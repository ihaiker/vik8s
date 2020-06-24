package cmd

import (
	"fmt"
	"github.com/ihaiker/vik8s/install/etcd"
	"github.com/ihaiker/vik8s/install/hosts"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/spf13/cobra"
	"time"
)

var etcdCmd = &cobra.Command{
	Use: "etcd", Short: "Install ETCD cluster",
	Long: `Install ETCD cluster. 
This program uses etcdadm for installation, for details https://github.com/kubernetes-sigs/etcdadm`,
}

func init() {
	etcdCmd.AddCommand(etcdInitCmd, etcdJoinCmd, etcdResetCmd)
}

var etcdInitCmd = &cobra.Command{
	Use: "init", Short: "Initialize a new etcd cluster",
	Args: cobra.MinimumNArgs(1),
	Example: `Args rule： [root:password@][ip-]ip[:port]
vik8s etcd init 172.16.100.11-172.16.100.13
vik8s etcd init 172.16.100.11 172.16.100.12 172.16.100.13`,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		ips := hosts.Add(etcd.Config.SSH, args...)
		etcd.InitCluster(ips[0])
		for _, ip := range ips[1:] {
			etcd.JoinCluster(ip)
		}
		fmt.Println("-=-=-=- SUCCESS -=-=-=-")
		return
	},
}

func init() {
	tenYear := time.Now().AddDate(44, 0, 0).Sub(time.Now())

	etcdInitCmd.Flags().StringVarP(&etcd.Config.SSH.Password, "ssh-passwd", "p", "", "password for ssh")
	etcdInitCmd.Flags().StringVar(&etcd.Config.SSH.PkFile, "ssh-pk", "$HOME/.ssh/id_rsa", "private key for ssh")
	etcdInitCmd.Flags().IntVarP(&etcd.Config.SSH.Port, "ssh-port", "P", 22, "default port for ssh")

	etcdInitCmd.Flags().StringVar(&etcd.Config.CertsDir, "certs-dir", "/etc/etcd/pki", "certificates directory")
	etcdInitCmd.Flags().DurationVar(&etcd.Config.CertsValidity, "certs-validity", tenYear, "Certificate validity time")

	etcdInitCmd.Flags().StringSliceVar(&etcd.Config.ServerCertExtraSans, "server-cert-extra-sans", []string{},
		"optional extra Subject Alternative Names for the etcd server signing cert, can be multiple comma separated DNS names or IPs")
	etcdInitCmd.Flags().StringVar(&etcd.Config.Snapshot, "snapshot", "", "Etcd v3 snapshot (local disk) file used to initialize member")
	etcdInitCmd.Flags().StringVar(&etcd.Config.RemoteSnapshot, "remote-snapshot", "", "Etcd v3 snapshot (remote disk at first node) file used to initialize member")
	etcdInitCmd.Flags().StringVar(&etcd.Config.Version, "version", "3.3.8", "etcd version")

	etcdInitCmd.Flags().StringVar(&etcd.Config.Source, "source", "", "the etcdadm source. if chain https://gitee.com/ihaiker/etcdadm else https://github.com/kubernetes-sigs/etcdadm")

	etcdInitCmd.Flags().SortFlags = false
}

var etcdJoinCmd = &cobra.Command{
	Use: "join", Short: "join nodes to etcd cluster",
	Example: `vik8s etcd join 172.16.100.10 172.16.100.11-172.16.100.13`,
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		etcd.Config.MustRead()
		ips := hosts.Add(etcd.Config.SSH, args...)
		for _, ip := range ips {
			etcd.JoinCluster(ip)
		}
		fmt.Println("-=-=-=- SUCCESS -=-=-=-")
	},
}

var etcdResetCmd = &cobra.Command{
	Use: "reset", Short: "reset etcd cluster",
	Example: `reset all: vik8s etcd reset
reset one node: vik8s etcd reset 172.16.100.10`,
	Run: func(cmd *cobra.Command, args []string) {
		etcd.Config.MustRead()
		ips := utils.ParseIPS(args)
		if len(ips) == 0 {
			ips = etcd.Config.Nodes
		}
		nodes := hosts.Add(etcd.Config.SSH, ips...)
		for _, node := range nodes {
			etcd.ResetCluster(node)
		}
		fmt.Println("-=-=-=- SUCCESS -=-=-=-")
	},
}
