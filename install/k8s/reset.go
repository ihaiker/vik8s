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
	err := node.Sudo().CmdStdout("kubeadm reset -f")
	if err != nil {
		node.Logger("reset %s", err.Error())
	}

	config.K8S().RemoveNode(node.Host)

	if len(config.K8S().Masters) == 0 && len(config.K8S().Nodes) == 0 {
		dataDir := paths.Join("kube")
		logs.Infof("remove data folder %s", dataDir)
		_ = os.RemoveAll(dataDir)
		if config.Config.ETCD != nil && len(config.Config.ETCD.Nodes) > 0 {
			logs.Infof("remove all cluster data in etcd")
			etcdNode := hosts.MustGet(config.Etcd().Nodes[0])
			err = etcdNode.Sudo().CmdPrefixStdout(etcd.Etcdctl("del /registry --prefix"))
			utils.Panic(err, "delete etcd cluster data /registry")
			err = etcdNode.Sudo().CmdPrefixStdout(etcd.Etcdctl("del /calico --prefix"))
			utils.Panic(err, "delete etcd cluster data /calico")
		}
	}

	logs.Infof("ipvsadm clear")
	if err = node.Sudo().Cmd("ipvsadm --clear"); err != nil {
		node.Logger("remove ipvsadm all role: %s", err.Error())
	}

	logs.Infof("clean CNI configuration")
	_ = node.Sudo().Cmd("rm -rf /etc/cni/net.d")
}
