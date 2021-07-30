package bases

import (
	"fmt"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"path/filepath"
)

func InstallTimeServices(node *ssh.Node, timezone string, timeServices ...string) {
	defer func() {
		EnableAndStartService("chronyd", true, node)
		node.MustCmd("chronyc -a makestep")
	}()
	node.MustCmd(fmt.Sprintf("rm -f /etc/localtime && cp -f %s /etc/localtime", filepath.Join("/usr/share/zoneinfo", timezone)))

	if node.Facts.MajorVersion == "7" {
		Install("chrony", "3.4", node) //fixbug 必须指定版本号，不然如何用户含有自己的repo会导致安装低版本出现问题
	} else {
		Install("chrony", "", node)
	}

	config := "allow all\n"
	for _, service := range timeServices {
		config += fmt.Sprintf("server %s iburst\n", service)
	}
	config += "\nlocal stratum 10\n"

	err := node.ScpContent([]byte(config), "/etc/chrony.conf")
	utils.Panic(err, "send ntp config")

	node.MustCmd(fmt.Sprintf("timedatectl set-timezone %s", timezone))
	node.MustCmd("timedatectl set-ntp true")
}
