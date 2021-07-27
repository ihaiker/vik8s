package bases

import (
	"fmt"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"strings"
)

func Install(mod, version string, node *ssh.Node) {
	node.Logger("install %s %s", mod, version)
	if node.IsCentOS() {
		installCentOS(mod, version, node)
	} else {
		installUbuntu(mod, version, node)
	}
}

func installUbuntu(mod, version string, node *ssh.Node) {
	installVersion, err := GetPackageVersion(node, mod)
	utils.Panic(err, "search package")

	if installVersion != "" {
		node.Logger("%s installed %s", mod, installVersion)
	}
	if (version != "" && installVersion == version) || (version == "" && installVersion != "") {
		return
	}

	if version == "" {
		err = node.SudoCmdWatcher(fmt.Sprintf("apt-get install %s", mod), utils.Stdout(node.Prefix()))
	} else {
		err = node.SudoCmdWatcher(fmt.Sprintf("apt-get install %s %s", mod, version), utils.Stdout(node.Prefix()))
	}
	utils.Panic(err, "install %s %s", mod, version)
}

func installCentOS(mod, version string, node *ssh.Node) {
	installVersion, err := GetPackageVersion(node, mod)
	utils.Panic(err, "search rpm version")

	if installVersion != "" {
		node.Logger("%s installed %s", mod, installVersion)
	}
	if (version != "" && installVersion == version) || (version == "" && installVersion != "") {
		return
	}

	if version == "" {
		err = node.SudoCmdWatcher(fmt.Sprintf("yum install -y %s", mod), utils.Stdout(node.Prefix()))
	} else {
		err = node.SudoCmdWatcher(fmt.Sprintf("yum install -y %s-%s", mod, version), utils.Stdout(node.Prefix()))
	}
	utils.Panic(err, "install package %s %s", mod, version)
}

func GetPackageVersion(node *ssh.Node, mod string) (version string, err error) {
	if node.IsCentOS() {
		version, err = node.SudoCmdString(fmt.Sprintf("rpm -qi %s | grep Version | awk '{printf $3}'", mod))
	} else {
		version, err = node.SudoCmdString(fmt.Sprintf("dpkg-query -s %s | grep Version | awk '{print $2}'", mod))
	}
	if err != nil && !strings.Contains(err.Error(), "not installed") {
		return
	}
	return
}

func Installs(node *ssh.Node, mods ...string) {
	for _, mod := range mods {
		Install(mod, "", node)
	}
}
