package k8s

import (
	"fmt"
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/install/bases"
	"github.com/ihaiker/vik8s/install/cri"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"gopkg.in/fatih/color.v1"
	"os"
	"strings"
)

func JoinControl(configure *config.Configuration, node *ssh.Node) {
	node.Logger("join control plane kubernetes cluster %s ", node.Host)

	if exists, _ := configure.K8S.ExistsNode(node.Host); exists {
		color.Red("%s already in the cluster\n", node.Host)
		return
	}
	master := configure.Hosts.MustGet(configure.K8S.Masters[0])
	bases.Check(node)
	bases.InstallTimeServices(node, configure.K8S.Timezone, configure.K8S.NTPServices...)
	cri.Install(configure, node)

	installKubernetesSoftware(configure, node)

	setNodeHosts(configure, node)
	setApiServerHosts(configure, node)
	setIpvsadmApiServer(configure, master, node)

	makeKubernetesCerts(configure, node)
	makeJoinControlPlaneConfigFiles(configure, node)

	remote := node.Vik8s("apply/kubeadm.yaml")
	bugfixImages(master, node, remote)

	joinCmd := getJoinCmd(master)
	control := fmt.Sprintf("%s --control-plane --apiserver-advertise-address=%s --ignore-preflight-errors=FileAvailable--etc-kubernetes-kubelet.conf --v=5", joinCmd, node.Host)
	utils.Panic(node.Sudo().CmdOutput(control, os.Stdout), "control plane join %s", node.Host)
	copyKubeletAdminConfig(node)

	fix(configure, master, node)

	configure.K8S.JoinNode(true, node.Host)
}

func JoinWorker(configure *config.Configuration, node *ssh.Node) {
	node.Logger("join worker kubernetes cluster %s ", node.Host)
	if exists, _ := configure.K8S.ExistsNode(node.Host); exists {
		color.Red("%s already in the cluster\n", node.Host)
		return
	}
	master := configure.Hosts.MustGet(configure.K8S.Masters[0])

	bases.Check(node)
	bases.InstallTimeServices(node, configure.K8S.Timezone, configure.K8S.NTPServices...)
	cri.Install(configure, node)

	installKubernetesSoftware(configure, node)

	setNodeHosts(configure, node)
	setApiServerHosts(configure, node)
	setIpvsadmApiServer(configure, master, node)

	makeWorkerConfigFiles(configure, node)

	remote := node.Vik8s("apply/kubeadm.yaml")
	bugfixImages(master, node, remote)

	joinCmd := getJoinCmd(master)
	cmd := fmt.Sprintf("%s --apiserver-advertise-address=%s --ignore-preflight-errors=FileAvailable--etc-kubernetes-kubelet.conf --v=5", joinCmd, node.Host)
	utils.Panic(node.Sudo().CmdOutput(cmd, os.Stdout), "join %s", node.Host)

	fix(configure, master, node)
	configure.K8S.JoinNode(false, node.Host)
}

func setNodeHosts(configure *config.Configuration, node *ssh.Node) {
	nodes := configure.Hosts.MustGets(append(configure.K8S.Masters, configure.K8S.Nodes...))
	setHosts(node, node.Host, node.Hostname)
	for _, n := range nodes {
		setHosts(n, node.Host, node.Hostname)
		setHosts(node, n.Host, n.Hostname)
	}
}

func setApiServerHosts(configure *config.Configuration, node *ssh.Node) {
	apiServerVip := configure.K8S.ApiServerVIP
	setHosts(node, apiServerVip, configure.K8S.ApiServer)
}

func setIpvsadmApiServer(configure *config.Configuration, master, node *ssh.Node) {
	apiServerVip := configure.K8S.ApiServerVIP
	_ = node.Sudo().Cmd(fmt.Sprintf("ipvsadm -D -t %s:6443", apiServerVip))

	err := node.Sudo().Cmd(fmt.Sprintf("ipvsadm -A -t %s:6443 -s rr", apiServerVip))
	utils.Panic(err, "add virtual-service")

	err = node.Sudo().Cmd(fmt.Sprintf("ipvsadm -a -t %s:6443 -r %s:6443 -m -w 1", apiServerVip, master.Host))
	utils.Panic(err, "add server-address to virtual-service")

	//fix 这个需要加入到开机启动项里面，不然会导致开机后ipvsadm丢失,
	err = node.Sudo().Cmd(`sed -i s/'IPVS_SAVE_ON_STOP="no"'/'IPVS_SAVE_ON_STOP="yes"'/g /etc/sysconfig/ipvsadm-config`)
	utils.Panic(err, "change ipvsadm-config")
	err = node.Sudo().Cmd(`sed -i s/'IPVS_SAVE_ON_RESTART="no"'/'IPVS_SAVE_ON_RESTART="yes"'/g /etc/sysconfig/ipvsadm-config`)
	utils.Panic(err, "change ipvsadm-config")
	err = node.Sudo().Cmd(`sh -c 'ipvsadm-save -n | sudo tee /etc/sysconfig/ipvsadm'`)
	utils.Panic(err, "change ipvsadm-config")
}

func fix(configure *config.Configuration, master, node *ssh.Node) {
	// for flannel
	//kubectl get nodes -o jsonpath='{.items[*].spec.podCIDR}'
	//kubectl get nodes -o template --template={{.spec.podCIDR}}
	err := master.Cmd(fmt.Sprintf("kubectl patch node %s -p '{\"spec\":{\"podCIDR\":\"%s\"}}'",
		node.Hostname, configure.K8S.PodCIDR))
	utils.Panic(err, "patch node %s %s", node.Hostname, configure.K8S.PodCIDR)
}

func getJoinCmd(node *ssh.Node) string {
	out, err := node.Sudo().CmdString("kubeadm token create --print-join-command")
	utils.Panic(err, "create cluster join token")
	return lastLine(out)
}

func lastLine(str string) string {
	str = strings.ReplaceAll(str, "\r", "")
	lines := strings.Split(str, "\n")
	return lines[len(lines)-1]
}
