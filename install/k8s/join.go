package k8s

import (
	"fmt"
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/install/bases"
	"github.com/ihaiker/vik8s/install/hosts"
	"github.com/ihaiker/vik8s/install/tools"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"gopkg.in/fatih/color.v1"
	"os"
	"strings"
)

func JoinControl(node *ssh.Node) {
	node.Logger("join control plane kubernetes cluster %s ", node.Host)

	if exists, _ := config.K8S().ExistsNode(node.Host); exists {
		color.Red("%s already in the cluster\n", node.Host)
		return
	}
	master := hosts.Get(config.K8S().Masters[0])
	installKubernetes(node)
	node.Logger("join control-plane")

	setNodeHosts(node)
	bases.InstallTimeServices(node, config.K8S().Timezone, config.Config.K8S.Masters...)
	setApiServerHosts(node)
	setIpvsadmApiServer(master, node)
	makeKubernetesCerts(node)
	makeJoinControlPlaneConfigFiles(node)

	joinCmd := getJoinCmd(master)
	control := fmt.Sprintf("%s --control-plane --apiserver-advertise-address=%s --ignore-preflight-errors=FileAvailable--etc-kubernetes-kubelet.conf", joinCmd, node.Host)
	utils.Panic(node.CmdStd(control, os.Stdout), "control plane join %s", node.Host)
	copyKubeletConfig(node)

	fix(master, node)

	config.K8S().JoinNode(true, node.Host)
}

func JoinWorker(node *ssh.Node) {
	node.Logger("join worker kubernetes cluster %s ", node.Host)
	if exists, _ := config.K8S().ExistsNode(node.Host); exists {
		color.Red("%s already in the cluster\n", node.Host)
		return
	}
	master := hosts.Get(config.K8S().Masters[0])
	installKubernetes(node)
	node.Logger("join worker")

	setNodeHosts(node)
	bases.InstallTimeServices(node, config.K8S().Timezone, config.Config.K8S.Masters...)
	setApiServerHosts(node)
	setIpvsadmApiServer(master, node)
	makeWorkerConfigFiles(node)

	joinCmd := getJoinCmd(master)
	cmd := fmt.Sprintf("%s --apiserver-advertise-address=%s --ignore-preflight-errors=FileAvailable--etc-kubernetes-kubelet.conf", joinCmd, node.Host)
	utils.Panic(node.CmdStd(cmd, os.Stdout), "join %s", node.Host)

	fix(master, node)
	config.K8S().JoinNode(false, node.Host)
}

func setNodeHosts(node *ssh.Node) {
	nodes := hosts.Gets(append(config.K8S().Masters, config.K8S().Nodes...))
	setHosts(node, node.Host, node.Hostname)
	for _, n := range nodes {
		setHosts(n, node.Host, node.Hostname)
		setHosts(node, n.Host, n.Hostname)
	}
}

func setApiServerHosts(node *ssh.Node) {
	apiServerVip := tools.GetVip(config.K8S().SvcCIDR, tools.Vik8sApiServer)
	setHosts(node, apiServerVip, config.K8S().ApiServer)
}

func setIpvsadmApiServer(master, node *ssh.Node) {
	apiServerVip := tools.GetVip(config.K8S().SvcCIDR, tools.Vik8sApiServer)
	_, _ = node.Cmd(fmt.Sprintf("ipvsadm -D -t %s:6443", apiServerVip))
	node.MustCmd(fmt.Sprintf("ipvsadm -A -t %s:6443 -s rr", apiServerVip))
	node.MustCmd(fmt.Sprintf("ipvsadm -a -t %s:6443 -r %s:6443 -m -w 1", apiServerVip, master.Host))

	//fix 这个需要加入到开机启动项里面，不然会导致开机后ipvsadm丢失,
	node.MustCmd(`sed -i s/'IPVS_SAVE_ON_STOP="no"'/'IPVS_SAVE_ON_STOP="yes"'/g /etc/sysconfig/ipvsadm-config`)
	node.MustCmd(`sed -i s/'IPVS_SAVE_ON_RESTART="no"'/'IPVS_SAVE_ON_RESTART="yes"'/g /etc/sysconfig/ipvsadm-config`)
	node.MustCmd(`ipvsadm-save -n > /etc/sysconfig/ipvsadm`)
}

func fix(master, node *ssh.Node) {
	// for flannel
	//kubectl get nodes -o jsonpath='{.items[*].spec.podCIDR}'
	//kubectl get nodes -o template --template={{.spec.podCIDR}}
	err := master.SudoCmdPrefixStdout(fmt.Sprintf("kubectl patch node %s -p '{\"spec\":{\"podCIDR\":\"%s\"}}'",
		node.Hostname, config.K8S().PodCIDR))
	utils.Panic(err, "patch node %s %s", node.Hostname, config.K8S().PodCIDR)
}

func getJoinCmd(node *ssh.Node) string {
	out, err := node.SudoCmdString("kubeadm token create --print-join-command")
	utils.Panic(err, "create cluster join token")
	return lastLine(out)
}

func lastLine(str string) string {
	str = strings.ReplaceAll(str, "\r", "")
	lines := strings.Split(str, "\n")
	return lines[len(lines)-1]
}
