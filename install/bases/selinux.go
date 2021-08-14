package bases

import (
	"github.com/ihaiker/vik8s/libs/ssh"
)

func disableSELinuxAndSwap(node *ssh.Node) {
	node.Logger("disable SELinux and swap")
	_ = node.Sudo().Cmd("setenforce 0")
	_ = node.Sudo().Cmd("swapoff -a")
	_ = node.Sudo().Cmd(`sed -i 's/.*swap.*//' /etc/fstab`)
}
