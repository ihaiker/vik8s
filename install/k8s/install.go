package k8s

import (
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/install/bases"
	"github.com/ihaiker/vik8s/install/repo"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
)

func installKubernetes(node *ssh.Node) {
	setRepo(node)
	sysctl(node)
	installKubeletAndKubeadm(node)
	modprobe(node)
}

func setRepo(node *ssh.Node) {
	node.Logger("set kubernetes repo")
	err := node.Sudo().ScpContent([]byte(repo.Kubernetes()), "/etc/yum.repos.d/kubernetes.repo")
	utils.Panic(err, "send /etc/yum.repos.d/kubernetes.repo")
}
func sysctl(node *ssh.Node) {
	//sysctl -w "net.netfilter.nf_conntrack_tcp_be_liberal=1"
	//https://juejin.cn/post/6976101827179708453
	err := node.Sudo().ScpContent([]byte(`
net.netfilter.nf_conntrack_tcp_be_liberal=1
net.bridge.bridge-nf-call-ip6tables=1
net.bridge.bridge-nf-call-iptables=1
net.ipv4.ip_forward=1
`), "/etc/sysctl.d/k8s.conf")
	utils.Panic(err, "send /etc/sysctl.d/k8s.conf")
	_ = node.Sudo().Cmd("sh -c 'echo 1 > /proc/sys/net/bridge/bridge-nf-call-iptables'")
	_ = node.Sudo().Cmd("sh -c 'echo 1 > /proc/sys/net/bridge/bridge-nf-call-ip6tables'")
	_ = node.Sudo().Cmd("sysctl -p")
	_ = node.Sudo().Cmd("update-alternatives --set iptables /usr/sbin/iptables-legacy")
	_ = node.Sudo().Cmd("update-alternatives --set ip6tables /usr/sbin/ip6tables-legacy")
	_ = node.Sudo().Cmd("update-alternatives --set arptables /usr/sbin/arptables-legacy")
	_ = node.Sudo().Cmd("update-alternatives --set ebtables /usr/sbin/ebtables-legacy")
}

func installKubeletAndKubeadm(node *ssh.Node) {
	version := config.K8S().Version[1:]
	node.Logger("Install kubelet & kubeadm v%s", version)

	bases.Install("ethtool", "", node)

	switch node.Facts.MajorVersion {
	case "7":
		bases.Install("ebtables", "", node)
	case "8":
		bases.Install("iptables-ebtables", "", node)
		bases.Install("iproute-tc", "", node)
	}

	bases.Install("bash-completion", "", node)
	bases.Install("ipvsadm", "", node)
	bases.Install("ipset", "", node)
	bases.Install("kubelet", version, node)
	bases.Install("kubeadm", version, node)

	_ = node.Sudo().Cmd("systemctl enable ipvsadm")
	_ = node.Sudo().Cmd("systemctl enable kubelet")
	_ = node.Sudo().Cmd("sh -c 'kubeadm completion bash > /etc/bash_completion.d/kubeadm'")
	_ = node.Sudo().Cmd("sh -c 'kubectl completion bash > /etc/bash_completion.d/kubectl'")
}

func modprobe(node *ssh.Node) {
	for _, mod := range []string{
		"ip_vs", "ip_vs_rr", "ip_vs_wrr", "ip_vs_sh", "ip_tables",
		"nf_conntrack", "br_netfilter", "dm_thin_pool",
	} {
		_ = node.Sudo().Cmd("modprobe " + mod)
	}
}
