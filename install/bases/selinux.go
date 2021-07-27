package bases

import (
	"github.com/ihaiker/vik8s/libs/ssh"
)

func disableSELinuxAndSwap(node *ssh.Node) {
	node.Logger("disable SELinux and swap")
	_ = node.SudoCmd("setenforce 0")
	_ = node.SudoCmd("swapoff -a")
	_ = node.SudoCmd(`sed -i 's/.*swap.*//' /etc/fstab`)
}
