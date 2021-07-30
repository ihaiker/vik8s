package bases

import "github.com/ihaiker/vik8s/libs/ssh"

func disableFirewalld(node *ssh.Node) {
	_ = node.SudoCmd("systemctl stop firewalld")
	_ = node.SudoCmd("systemctl disable firewalld")
	_ = node.SudoCmdStdout("systemctl stop iptables")
	_ = node.SudoCmdStdout("systemctl disable iptables")
}
