package bases

import "github.com/ihaiker/vik8s/libs/ssh"

func disableFirewalld(node *ssh.Node) {
	_ = node.Sudo().Cmd("systemctl stop firewalld")
	_ = node.Sudo().Cmd("systemctl disable firewalld")
	_ = node.Sudo().CmdStdout("systemctl stop iptables")
	_ = node.Sudo().CmdStdout("systemctl disable iptables")
}
