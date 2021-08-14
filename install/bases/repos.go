package bases

import (
	"fmt"
	"github.com/ihaiker/vik8s/install/paths"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
)

func setAliRepo(node *ssh.Node) {
	if node.IsUbuntu() {
		setAliRepoUbuntu(node)
	} else {
		setAliRepoCentOS(node)
	}
}

//添加一个Repo文件
func AddRepoFile(node *ssh.Node, name string, content []byte) {
	var remoteRepoPath string
	if node.IsCentOS() {
		remoteRepoPath = fmt.Sprintf("/etc/yum.repos.d/%s.repo", name)
	} else {
		remoteRepoPath = fmt.Sprintf("/etc/apt/sources.list.d/%s.list", name)
	}
	utils.Panic(node.Sudo().ScpContent(content, remoteRepoPath), "scp repo content")
}

func setAliRepoUbuntu(node *ssh.Node) {
	sourceList := []byte(`
deb http://mirrors.aliyun.com/ubuntu/ xenial main restricted universe multiverse  
deb http://mirrors.aliyun.com/ubuntu/ xenial-security main restricted universe multiverse  
deb http://mirrors.aliyun.com/ubuntu/ xenial-updates main restricted universe multiverse  
deb http://mirrors.aliyun.com/ubuntu/ xenial-backports main restricted universe multiverse  
deb http://mirrors.aliyun.com/ubuntu/ xenial-proposed main restricted universe multiverse  
deb-src http://mirrors.aliyun.com/ubuntu/ xenial main restricted universe multiverse  
deb-src http://mirrors.aliyun.com/ubuntu/ xenial-security main restricted universe multiverse  
deb-src http://mirrors.aliyun.com/ubuntu/ xenial-updates main restricted universe multiverse  
deb-src http://mirrors.aliyun.com/ubuntu/ xenial-backports main restricted universe multiverse  
deb-src http://mirrors.aliyun.com/ubuntu/ xenial-proposed main restricted universe multiverse  
`)
	err := node.Sudo().ScpContent(sourceList, "/etc/apt/sources.list.d/vik8s.list")
	utils.Panic(err, "add sources list")

	err = node.Sudo().CmdWatcher("apt-get update", utils.Stdout(node.Prefix()))
	utils.Panic(err, "update source")
}

func setAliRepoCentOS(node *ssh.Node) {
	Installs(node, "epel-release")
	if paths.China {
		repoUrl := fmt.Sprintf("http://mirrors.aliyun.com/repo/Centos-%s.repo", node.Facts.MajorVersion)
		err := node.Sudo().Cmd("curl --silent -o /etc/yum.repos.d/vik8s.repo " + repoUrl)
		utils.Panic(err, "download AliCloud repos")

		if node.Facts.MajorVersion == "7" {
			err = node.Sudo().Cmd("curl --silent -o /etc/yum.repos.d/epel.repo http://mirrors.aliyun.com/repo/epel-7.repo")
			utils.Panic(err, "download AliCloud repos")
		}
	}
	Installs(node, "yum-utils", "lvm2", "device-mapper-persistent-data")
	_ = node.Sudo().Cmd("yum makecache")
}
