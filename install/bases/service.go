package bases

import (
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
)

func EnableAndStartService(name string, mustRestart bool, node *ssh.Node) {
	node.Logger("start service %s", name)
	status, _ := node.Sudo().HideLog().CmdString("systemctl status " + name + " | grep 'Active:' | awk '{printf $2}'")
	node.Logger("the service %s status is: ", name, status)
	if status == "inactive" {
		_ = node.Sudo().Cmd("systemctl enable " + name)
	}
	if status == "active" && mustRestart {
		err := node.Sudo().Cmd("systemctl stop " + name)
		utils.Panic(err, "stop service %s", name)
	}
	status, _ = node.Sudo().HideLog().CmdString("systemctl status " + name + " | grep 'Active:' | awk '{printf $2}'")
	if status != "active" {
		err := node.Sudo().Cmd("systemctl start " + name)
		utils.Panic(err, "start service %s at node(%s)", name, node.Hostname)
	}
}
