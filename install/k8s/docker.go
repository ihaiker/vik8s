package k8s

import (
	"encoding/json"
	"fmt"
	"github.com/ihaiker/vik8s/install/repo"
	"github.com/ihaiker/vik8s/install/tools"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"strings"
)

func daemonJson() map[string]interface{} {
	daemon := map[string]interface{}{
		"exec-opts": []string{"native.cgroupdriver=systemd"},
		"hosts": []string{
			"fd://",
			"tcp://0.0.0.0:2375",
			"unix:///var/run/docker.sock",
		},
		"registry-mirrors": []string{},
	}
	if tools.China {
		daemon["registry-mirrors"] = append(daemon["registry-mirrors"].([]string), []string{
			"https://dockerhub.azk8s.cn",
			"https://docker.mirrors.ustc.edu.cn",
			"http://hub-mirror.c.163.com",
			"https://registry.cn-hangzhou.aliyuncs.com",
		}...)
	}
	daemon["insecure-registries"] = Config.Docker.Registry
	daemon["registry-mirrors"] = append(daemon["registry-mirrors"].([]string), Config.Docker.Registry...)
	return daemon
}

func checkDocker(node *ssh.Node) {
	defer func() {
		if e := recover(); e == nil {
			tools.EnableAndStartService("docker", node)
			//TOME 如果 Node 上安装的 Docker 版本大于 1.12，那么 Docker 会把默认的 iptables FORWARD 策略改为 DROP。转发丢弃
			//这会引发 Pod 网络访问的问题
			node.MustCmd("iptables -P FORWARD ACCEPT")
		} else {
			panic(e) //继续向下抛
		}
	}()

	dockerVersion := node.MustCmd2String("rpm -qi docker-ce | grep Version | awk '{printf $3}'")
	if dockerVersion != "" && (dockerVersion == Config.Docker.Version || !Config.Docker.CheckVersion) {
		node.Logger("docker installd %s", dockerVersion)
	} else {
		node.Logger("install docker %s", dockerVersion)

		tools.AddRepo(repo.Docker(), node)
		//install containerd.io
		//TOME 多版本问题，docker镜像里面的最高版本1.2.0，在docker19.+最低版本要求 1.2.2
		if node.ReleaseName == "CentOS" && node.MajorVersion == "8" {
			node.Logger("CentOS 8 check container.io")
			//tools.Install("containerd.io", "1.2.10", node) 版本不在docker镜像里面
			containerIO := node.MustCmd2String("rpm -qa | grep containerd.io || echo NOT_FOUND")
			if containerIO == "NOT_FOUND" || utils.VersionCompose(strings.Split(containerIO, "-")[1], "1.2.3") < 0 {
				_, _ = node.Cmd("yum remove -y containerd.io")
				node.Logger("install containerd.io")
				_, err := node.Cmd(fmt.Sprintf("yum install -y %s", repo.Containerd()))
				utils.Panic(err, "install containerd.io")
			}
		}

		tools.Install("docker-ce", Config.Docker.Version, node)
		tools.Install("docker-ce-cli", Config.Docker.Version, node)
	}

	//set docker daemon.json
	_ = node.MustCmd2String("mkdir -p /etc/docker")
	if Config.Docker.DaemonJson != "" {
		err := node.Scp(Config.Docker.DaemonJson, "/etc/docker/daemin.json")
		utils.Panic(err, "scp daemon.json")
	} else {
		bs, _ := json.MarshalIndent(daemonJson(), "", "    ")
		err := node.ScpContent(bs, "/etc/docker/daemon.json")
		utils.Panic(err, "scp daemon.json")
	}

	// set docker.service
	_ = node.MustCmd2String("sed -i 's/-H fd:\\/\\///g' /usr/lib/systemd/system/docker.service ")
}
