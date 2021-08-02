package k8s

import (
	"fmt"
	"github.com/ihaiker/vik8s/install/bases"
	"github.com/ihaiker/vik8s/install/hosts"
	"github.com/ihaiker/vik8s/install/paths"
	"github.com/ihaiker/vik8s/install/tools"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/reduce"
	yamls "github.com/ihaiker/vik8s/yaml"
	"io/ioutil"
	"os"
)

func InitCluster(node *ssh.Node) *ssh.Node {
	utils.Line("init kubernetes cluster %s", node.Host)

	preCheck(node)
	//fix 这里需要先行加入，因为在初始化模板文件中需要使用
	Config.JoinNode(true, node.Host)

	node.Logger("init cluster")
	{
		setHosts(node, node.Host, Config.Kubernetes.ApiServer)
		setHosts(node, node.Host, node.Hostname)
		bases.InstallTimeServices(node, Config.Timezone, Config.NTPServices...)
		makeCerts(node)
		makeJoinControlPlaneConfigFiles(node)
		initKubernetes(node)
		applyApiServerEndpoint(node)
	}
	return node
}

func ResetNode(node *ssh.Node) {
	_, _ = node.Cmd("kubeadm reset -f")
	Config.RemoveNode(node.Host)
	if len(Config.Masters) == 0 && len(Config.Nodes) == 0 {
		_ = os.RemoveAll(paths.Join("kube"))
		//TODO 这里设置
		/*if err := etcd.Config.Read(); err == nil {
			_, _ = hosts.Get(etcd.Config.Nodes[0]).Cmd("etcdctl.sh del /registry --prefix")
			_, _ = hosts.Get(etcd.Config.Nodes[0]).Cmd("etcdctl.sh del /calico --prefix")
		}*/
	}
}

func setHosts(node *ssh.Node, ip, domain string) {
	_ = node.MustCmd2String(fmt.Sprintf("sed -i /%s/d /etc/hosts", domain))
	_ = node.MustCmd2String(fmt.Sprintf("echo '%s %s' >> /etc/hosts", ip, domain))
}

func scpKubeConfig(node *ssh.Node) string {
	kubeadmConfigPath := string(yamls.MustAsset("yaml/kubeadm-yaml"))

	if Config.Kubernetes.KubeadmConfig != "" {
		configBytes, err := ioutil.ReadFile(Config.Kubernetes.KubeadmConfig)
		utils.Panic(err, "read kubeadm-config file %s", Config.Kubernetes.KubeadmConfig)
		kubeadmConfigPath = string(configBytes)
	}

	remote := node.Vik8s("apply/kubeadm-yaml")
	node.Logger("scp kubeadm.yaml %s", remote)

	data := paths.Json{
		"Etcd":    Config.ETCD,
		"Masters": hosts.Gets(Config.Masters), "Workers": hosts.Gets(Config.Nodes),
		"Nodes":   hosts.Gets(Config.AllNode()),
		"Kubeadm": Config.Kubernetes,
	}
	kubeadmConfig := tools.Template(kubeadmConfigPath, data)
	err := node.ScpContent(kubeadmConfig.Bytes(), remote)
	utils.Panic(err, "scp kubeadm-config file")
	return remote
}

func initKubernetes(node *ssh.Node) {
	remote := scpKubeConfig(node)
	err := node.CmdStd(fmt.Sprintf("kubeadm init --config=%s --upload-certs", remote), os.Stdout)
	utils.Panic(err, "kubeadm init")
	copyKubeletConfg(node)
}

func copyKubeletConfg(node *ssh.Node) {
	kubeDir := node.HomeDir(".kube")
	kubeConfig := node.HomeDir(".kube/config")
	_ = node.MustCmd2String(fmt.Sprintf("mkdir -p %s  && cp -f /etc/kubernetes/admin.conf %s", kubeDir, kubeConfig))
}

func applyApiServerEndpoint(node *ssh.Node) {
	name := "yaml/vik8s-api-server.conf"
	data := paths.Json{
		"Etcd":    Config.ETCD,
		"Masters": hosts.Gets(Config.Masters), "Workers": hosts.Gets(Config.Nodes),
		"Nodes":   hosts.Gets(Config.AllNode()),
		"Kubeadm": Config.Kubernetes,
	}
	//tools.MustScpAndApplyAssert(node, name, data)
	reduce.MustApplyAssert(node, name, data)
}
