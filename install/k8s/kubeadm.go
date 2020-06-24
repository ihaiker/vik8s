package k8s

import (
	"github.com/ihaiker/vik8s/install/repo"
	"github.com/ihaiker/vik8s/install/tools"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
)

func checkKubernetes(node *ssh.Node) {
	setRepo(node)
	sysctl(node)
	installKubeletAndKubeadm(node)
	modprobe(node)
}

func setRepo(node *ssh.Node) {
	utils.Line("set kubernetes repo")
	err := node.ScpContent([]byte(repo.Kubernetes()), "/etc/yum.repos.d/kubernetes.repo")
	utils.Panic(err, "send /etc/yum.repos.d/kubernetes.repo")
}
func sysctl(node *ssh.Node) {
	err := node.ScpContent([]byte(`
net.bridge.bridge-nf-call-ip6tables=1
net.bridge.bridge-nf-call-iptables=1
net.ipv4.ip_forward=1
`), "/etc/sysctl.d/k8s.conf")
	utils.Panic(err, "send /etc/sysctl.d/k8s.conf")
	node.MustCmd("echo 1 > /proc/sys/net/bridge/bridge-nf-call-iptables")
	node.MustCmd("echo 1 > /proc/sys/net/bridge/bridge-nf-call-ip6tables")
	_ = node.MustCmd2String("sysctl -p")
	_, _ = node.Cmd("update-alternatives --set iptables /usr/sbin/iptables-legacy")
	_, _ = node.Cmd("update-alternatives --set ip6tables /usr/sbin/ip6tables-legacy")
	_, _ = node.Cmd("update-alternatives --set arptables /usr/sbin/arptables-legacy")
	_, _ = node.Cmd("update-alternatives --set ebtables /usr/sbin/ebtables-legacy")
}

func installKubeletAndKubeadm(node *ssh.Node) {
	utils.Line("Install kubelet & kubeadm")

	tools.Install("ethtool", "", node)

	switch node.MajorVersion {
	case "7":
		tools.Install("ebtables", "", node)
	case "8":
		tools.Install("iptables-ebtables", "", node)
		tools.Install("iproute-tc", "", node)
	}

	tools.Install("bash-completion", "", node)
	tools.Install("ipvsadm", "", node)
	tools.Install("ipset", "", node)
	tools.Install("kubelet", Config.Kubernetes.Version, node)
	tools.Install("kubeadm", Config.Kubernetes.Version, node)

	_, _ = node.Cmd("systemctl enable ipvsadm")
	_, _ = node.Cmd("systemctl enable kubelet")
	_, _ = node.Cmd("kubeadm completion bash > /etc/bash_completion.d/kubeadm")
	_, _ = node.Cmd("kubectl completion bash > /etc/bash_completion.d/kubectl")
}

func modprobe(node *ssh.Node) {
	for _, mod := range []string{
		"ip_vs", "ip_vs_rr", "ip_vs_wrr", "ip_vs_sh", "ip_tables",
		"nf_conntrack", "br_netfilter", "dm_thin_pool",
	} {
		_ = node.MustCmd2String("modprobe " + mod)
	}
}
