package k8s

import (
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/install/etcd"
	"github.com/ihaiker/vik8s/install/hosts"
	"github.com/ihaiker/vik8s/install/paths"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"os"
)

func ResetNode(node *ssh.Node) {
	err := node.SudoCmdPrefixStdout("kubeadm reset -f")
	utils.Panic(err, "kubernetes cluster reset")

	config.K8S().RemoveNode(node.Host)
	if len(config.K8S().Masters) == 0 && len(config.K8S().Nodes) == 0 {
		_ = os.RemoveAll(paths.Join("kube"))
		if config.Config.ETCD != nil && len(config.Config.ETCD.Nodes) > 0 {
			etcdNode := hosts.Get(config.Etcd().Nodes[0])
			err = etcdNode.SudoCmdPrefixStdout(etcd.Etcdctl("del /registry --prefix"))
			utils.Panic(err, "delete etcd cluster data /registry")
			err = etcdNode.SudoCmdPrefixStdout(etcd.Etcdctl("del /calico --prefix"))
			utils.Panic(err, "delete etcd cluster data /calico")
		}
	}
}
