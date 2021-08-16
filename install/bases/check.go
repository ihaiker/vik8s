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

func InstallJQYQTools(node *ssh.Node) {
	Install("jq", "", node)

	downloadUrl, err := node.CmdString("curl https://api.github.com/repos/mikefarah/yq/releases/latest |" +
		" jq  -r '.assets[] | select(.name == \"yq_linux_amd64\") | .browser_download_url'")
	utils.Panic(err, "get yq download url error")

	err = node.Sudo().CmdStdout("curl -o /usr/local/bin/yq " + downloadUrl)
	utils.Panic(err, "download yq")

	err = node.Sudo().Cmd("chmod +x /usr/local/bin/yq")
	utils.Panic(err, "change yq mode")
}
