package k8s

import (
	"fmt"
	"github.com/ihaiker/vik8s/install"
	"github.com/ihaiker/vik8s/install/hosts"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"gopkg.in/fatih/color.v1"
	"os"
	"strings"
)

func JoinControl(node *ssh.Node) {
	utils.Line("join control plane kubernetes cluster %s ", node.Host)

	if exists, _ := Config.ExistsNode(node.Host); exists {
		color.Red("%s already in the cluster\n", node.Host)
		return
	}
	master := hosts.Get(Config.Masters[0])
	preCheck(node)
	node.Logger("join control-plane")

	setNodeHosts(node)
	install.InstallChronyServices(node, Config.Timezone, master.Host)
	setApiServerHosts(node)
	setIpvsadmApiServer(master, node)
	makeCerts(node)
	makeJoinControlPlaneConfigFiles(node)

	joinCmd := getJoinCmd(master)
	control := fmt.Sprintf("%s --control-plane --apiserver-advertise-address=%s --ignore-preflight-errors=FileAvailable--etc-kubernetes-kubelet.conf", joinCmd, node.Host)
	utils.Panic(node.CmdStd(control, os.Stdout), "control plane join %s", node.Host)
	copyKubeletConfg(node)

	fix(master, node)

	Config.JoinNode(true, node.Host)
}

func JoinWorker(node *ssh.Node) {
	utils.Line("join worker kubernetes cluster %s ", node.Host)
	if exists, _ := Config.ExistsNode(node.Host); exists {
		color.Red("%s already in the cluster\n", node.Host)
		return
	}
	master := hosts.Get(Config.Masters[0])
	preCheck(node)
	node.Logger("join worker")

	setNodeHosts(node)
	install.InstallChronyServices(node, Config.Timezone, master.Host)
	setApiServerHosts(node)
	setIpvsadmApiServer(master, node)
	makeWorkerConfigFiles(node)

	joinCmd := getJoinCmd(master)
	cmd := fmt.Sprintf("%s --apiserver-advertise-address=%s --ignore-preflight-errors=FileAvailable--etc-kubernetes-kubelet.conf", joinCmd, node.Host)
	utils.Panic(node.CmdStd(cmd, os.Stdout), "join %s", node.Host)

	fix(master, node)
	Config.JoinNode(false, node.Host)
}

func setNodeHosts(node *ssh.Node) {
	nodes := hosts.Gets(Config.AllNode())
	setHosts(node, node.Host, node.Hostname)
	for _, n := range nodes {
		setHosts(n, node.Host, node.Hostname)
		setHosts(node, n.Host, n.Hostname)
	}
}

func setApiServerHosts(node *ssh.Node) {
	setHosts(node, Config.Kubernetes.ApiServerVIP, Config.Kubernetes.ApiServer)
}

func setIpvsadmApiServer(master, node *ssh.Node) {
	_, _ = node.Cmd(fmt.Sprintf("ipvsadm -D -t %s:6443", Config.Kubernetes.ApiServerVIP))
	node.MustCmd(fmt.Sprintf("ipvsadm -A -t %s:6443 -s rr", Config.Kubernetes.ApiServerVIP))
	node.MustCmd(fmt.Sprintf("ipvsadm -a -t %s:6443 -r %s:6443 -m -w 1", Config.Kubernetes.ApiServerVIP, master.Host))

	//fix 这个需要加入到开机启动项里面，不然会导致开机后ipvsadm丢失,
	node.MustCmd(`sed -i s/'IPVS_SAVE_ON_STOP="no"'/'IPVS_SAVE_ON_STOP="yes"'/g /etc/sysconfig/ipvsadm-config`)
	node.MustCmd(`sed -i s/'IPVS_SAVE_ON_RESTART="no"'/'IPVS_SAVE_ON_RESTART="yes"'/g /etc/sysconfig/ipvsadm-config`)
	node.MustCmd(`ipvsadm-save -n > /etc/sysconfig/ipvsadm`)
}

func fix(master, node *ssh.Node) {
	// for flannel
	//kubectl get nodes -o jsonpath='{.items[*].spec.podCIDR}'
	//kubectl get nodes -o template --template={{.spec.podCIDR}}
	_, _ = master.Cmd(fmt.Sprintf("kubectl patch node %s -p '{\"spec\":{\"podCIDR\":\"%s\"}}'", node.Hostname, Config.Kubernetes.PodCIDR))

}

func getJoinCmd(node *ssh.Node) string {
	return lastLine(node.MustCmd2String("kubeadm token create --print-join-command"))
}

func lastLine(str string) string {
	str = strings.ReplaceAll(str, "\r", "")
	lines := strings.Split(str, "\n")
	return lines[len(lines)-1]
}
