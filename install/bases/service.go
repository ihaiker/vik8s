package bases

import (
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
)

func EnableAndStartService(name string, mustRestart bool, node *ssh.Node) {
	status, _ := node.Sudo().CmdString("systemctl status " + name + " | grep 'Active:' | awk '{printf $2}'")
	if status == "inactive" {
		_ = node.Sudo().Cmd("systemctl enable " + name)
	}
	if status == "active" && mustRestart {
		err := node.Sudo().Cmd("systemctl stop " + name)
		utils.Panic(err, "stop service %s", name)
	}
	err := node.Sudo().Cmd("systemctl start " + name)
	utils.Panic(err, "start service %s at node(%s)", name, node.Hostname)
}
