package k8s

import (
	"github.com/ihaiker/vik8s/install"
	"github.com/ihaiker/vik8s/libs/ssh"
)

func preCheck(node *ssh.Node) {
	install.PreCheck(node)
	disableFireWalld(node)
	checkDocker(node)
	checkKubernetes(node)
}

func disableFireWalld(node *ssh.Node) {
	_, _ = node.Cmd("systemctl stop firewalld")
	_, _ = node.Cmd("systemctl disable firewalld")
	_, _ = node.Cmd("systemctl stop iptables")
	_, _ = node.Cmd("systemctl disable iptables")
}
