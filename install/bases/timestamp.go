package bases

import (
	"fmt"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"path/filepath"
)

func InstallTimeServices(node *ssh.Node, timezone string, timeServices ...string) {
	err := node.SudoCmd("rm -f /etc/localtime")
	utils.Panic(err, "set timezone")

	err = node.SudoCmd(fmt.Sprintf("cp -f %s /etc/localtime",
		filepath.Join("/usr/share/zoneinfo", timezone)))
	utils.Panic(err, "set timezone")

	if node.Facts.MajorVersion == "7" {
		//fixbug 必须指定版本号，不然如何用户含有自己的repo会导致安装低版本出现问题
		Install("chrony", "3.4", node)
	} else {
		Install("chrony", "", node)
	}

	config := "allow all\n"
	for _, service := range timeServices {
		config += fmt.Sprintf("server %s iburst\n", service)
	}
	config += "\nlocal stratum 10\n"

	err = node.SudoScpContent([]byte(config), "/etc/chrony.conf")
	utils.Panic(err, "send ntp config")

	err = node.SudoCmd(fmt.Sprintf("timedatectl set-timezone %s", timezone))
	utils.Panic(err, "set timezone")

	err = node.SudoCmd("timedatectl set-ntp true")
	utils.Panic(err, "set timezone")

	err = node.SudoCmd("chronyc -a makestep")
	utils.Panic(err, "set timezone")

	EnableAndStartService("chronyd", true, node)
}
