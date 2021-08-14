package k8s

import (
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/install/etcd"
	"github.com/ihaiker/vik8s/install/hosts"
	"github.com/ihaiker/vik8s/install/paths"
	"github.com/ihaiker/vik8s/libs/logs"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"os"
)

func ResetNode(node *ssh.Node) {
	err := node.SudoCmdStdout("kubeadm reset -f")
	utils.Panic(err, "kubernetes cluster reset")

	if config.K8S() != nil {
		config.K8S().RemoveNode(node.Host)
	}

	if config.K8S() != nil && len(config.K8S().Masters) == 0 && len(config.K8S().Nodes) == 0 {
		dataDir := paths.Join("kube")
		logs.Infof("remove data folder %s", dataDir)
		_ = os.RemoveAll(dataDir)
		if config.Config.ETCD != nil && len(config.Config.ETCD.Nodes) > 0 {
			logs.Infof("remove all cluster data in etcd")
			etcdNode := hosts.Get(config.Etcd().Nodes[0])
			err = etcdNode.SudoCmdPrefixStdout(etcd.Etcdctl("del /registry --prefix"))
			utils.Panic(err, "delete etcd cluster data /registry")
			err = etcdNode.SudoCmdPrefixStdout(etcd.Etcdctl("del /calico --prefix"))
			utils.Panic(err, "delete etcd cluster data /calico")
		}
	}

	logs.Infof("ipvsadm clear")
	err = node.SudoCmd("ipvsadm --clear")
	utils.Panic(err, "remove ipvsadm all role")

	logs.Infof("clean CNI configuration")
	_ = node.SudoCmd("rm -rf /etc/cni/net.d")
}
