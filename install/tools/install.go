package tools

import (
	"fmt"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
)

func Install(mod, version string, node *ssh.Node) {
	installVersion := node.MustCmd2String(fmt.Sprintf("rpm -qi %s | grep Version | awk '{printf $3}'", mod))
	if installVersion != "" {
		node.Logger("%s installed %s", mod, installVersion)
	}
	if (version != "" && installVersion == version) || (version == "" && installVersion != "") {
		return
	}
	var err error
	if version == "" {
		err = node.CmdChannel(fmt.Sprintf("yum install -y %s", mod), utils.Stdout(node.Hostname))
	} else {
		err = node.CmdChannel(fmt.Sprintf("yum install -y %s-%s", mod, version), utils.Stdout(node.Hostname))
	}
	utils.Panic(err, "install %s %s", mod, version)
}

func Installs(node *ssh.Node, mods ...string) {
	for _, mod := range mods {
		Install(mod, "", node)
	}
}

func AddRepo(url string, node *ssh.Node) {
	_, _ = node.Cmd(fmt.Sprintf("yum-config-manager --add-repo %s", url))
}

//添加一个Repo文件
func AddRepoFile(name, content string, node *ssh.Node) {
	remoteRepoPath := fmt.Sprintf("/etc/yum.repos.d/%s.repo", name)
	node.MustScpContent([]byte(content), remoteRepoPath)
}

func EnableAndStartService(name string, node *ssh.Node) {
	status := node.MustCmd2String("systemctl status " + name + " | grep 'Active:' | awk '{printf $2}'")
	if status == "inactive" {
		_, _ = node.Cmd("systemctl enable " + name)
	}
	//_ = node.MustCmd2String("systemctl restart " + name)
	_, _ = node.Cmd("systemctl stop " + name)
	node.MustCmd("systemctl start " + name)
}
