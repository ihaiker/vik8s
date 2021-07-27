package bases

import (
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
)

func EnableAndStartService(name string, node *ssh.Node) {
	status, _ := node.SudoCmdString("systemctl status " + name + " | grep 'Active:' | awk '{printf $2}'")
	if status == "inactive" {
		_, _ = node.Cmd("systemctl enable " + name)
	}
	//_ = node.MustCmd2String("systemctl restart " + name)
	_ = node.SudoCmd("systemctl stop " + name)
	err := node.SudoCmd("systemctl start " + name)
	utils.Panic(err, "start service %s at node(%s)", name, node.Hostname)
}
