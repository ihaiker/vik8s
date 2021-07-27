package bases

import "github.com/ihaiker/vik8s/libs/ssh"

func disableFirewalld(node *ssh.Node) {
	_ = node.SudoCmd("systemctl stop firewalld")
	_ = node.SudoCmd("systemctl disable firewalld")
	_ = node.SudoCmd("systemctl stop iptables")
	_ = node.SudoCmd("systemctl disable iptables")
}
