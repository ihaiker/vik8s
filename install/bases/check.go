package bases

import (
	"github.com/hashicorp/go-version"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
)

func Check(node *ssh.Node) {
	checkDistribution(node)
	setAliRepo(node)
	disableSELinuxAndSwap(node)
	disableFirewalld(node)
}

func checkDistribution(node *ssh.Node) {
	node.Logger("check distribution")
	v1, _ := version.NewVersion("4.1")
	v2, _ := version.NewVersion(node.Facts.KernelVersion)
	utils.Assert(v1.LessThanOrEqual(v2), "%s The kernel version is too low, please upgrade the kernel first, "+
		"your current version is: %s, the minimum requirement is %s", node.Prefix(), v2.String(), v1.String())
}
