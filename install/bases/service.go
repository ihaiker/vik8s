package bases

import (
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
)

func EnableAndStartService(name string, mustRestart bool, node *ssh.Node) {
	status, _ := node.SudoCmdString("systemctl status " + name + " | grep 'Active:' | awk '{printf $2}'")
	if status == "inactive" {
		_ = node.SudoCmd("systemctl enable " + name)
	}
	if status == "active" && mustRestart {
		err := node.SudoCmd("systemctl stop " + name)
		utils.Panic(err, "stop service %s", name)
	}
	err := node.SudoCmd("systemctl start " + name)
	utils.Panic(err, "start service %s at node(%s)", name, node.Hostname)
}
