package install

import (
	"fmt"
	"github.com/hashicorp/go-version"
	"github.com/ihaiker/vik8s/install/hosts"
	"github.com/ihaiker/vik8s/install/tools"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"path/filepath"
)

func PreCheck(node *ssh.Node) {
	setAliRepo(node)
	checkDistribution(node)
	disableSELinuxAndSwap(node)
	disableFireWalld(node)
}

func disableFireWalld(node *ssh.Node) {
	_, _ = node.Cmd("systemctl stop firewalld")
	_, _ = node.Cmd("systemctl disable firewalld")
	_, _ = node.Cmd("systemctl stop iptables")
	_, _ = node.Cmd("systemctl disable iptables")
}

//经检查系统类型是否满足
func support(node *ssh.Node) {
	ss := []string{
		"CentOS 7",
		"CentOS 8",
	}
	utils.Assert(node.ReleaseName != "unsupport", "unsupport system")
	for _, s := range ss {
		if fmt.Sprintf("%s %s", node.ReleaseName, node.MajorVersion) == s {
			return
		}
	}
	node.Logger("[warn] Unstrictly tested system")
}

func setAliRepo(node *ssh.Node) {
	tools.Installs(node, "epel-release")

	if tools.China {
		repoUrl := fmt.Sprintf("http://mirrors.aliyun.com/repo/Centos-%s.repo", node.MajorVersion)
		node.MustCmd("curl --silent -o /etc/yum.repos.d/CentOS-vik8s.repo " + repoUrl)
		if node.MajorVersion == "7" {
			node.MustCmd("curl --silent -o /etc/yum.repos.d/epel.repo http://mirrors.aliyun.com/repo/epel-7.repo")
		}
	}
	tools.Installs(node, "yum-utils", "lvm2", "device-mapper-persistent-data")
}

func checkDistribution(node *ssh.Node) {
	v1, _ := version.NewVersion("4.1")
	support(node)
	v2, _ := version.NewVersion(node.KernelVersion)
	if v1.GreaterThanOrEqual(v2) {
		hosts.Remove(node.Hostname) //fixbug: 如果版本错误，删除本地管理，不然没办法安装
	}
	utils.Assert(v1.LessThanOrEqual(v2), "[%s,%s] The kernel version is too low, please upgrade the kernel first, "+
		"your current version is: %s, the minimum requirement is %s", node.Address(), node.Hostname, v2.String(), v1.String())
}

func disableSELinuxAndSwap(node *ssh.Node) {
	utils.Line("disable SELinux and swap")
	_, _ = node.Cmd("setenforce 0")
	_, _ = node.Cmd("swapoff -a")
	_, _ = node.Cmd(`sed -i 's/.*swap.*//' /etc/fstab`)
}

func InstallChronyServices(node *ssh.Node, timezone string, timeServices ...string) {
	defer func() {
		tools.EnableAndStartService("chronyd", node)
		node.MustCmd("chronyc -a makestep")
	}()

	node.MustCmd(fmt.Sprintf("rm -f /etc/localtime && cp -f %s /etc/localtime", filepath.Join("/usr/share/zoneinfo", timezone)))
	tools.Install("chrony", "3.4", node) //fixbug 必须指定版本号，不然如何用户含有自己的repo会导致安装低版本出现问题

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
