package cri

import (
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/install/cri/containerd"
	"github.com/ihaiker/vik8s/install/cri/docker"
	"github.com/ihaiker/vik8s/install/paths"
	"github.com/ihaiker/vik8s/libs/ssh"
)

func Install(configure *config.Configuration, node *ssh.Node) {
	if configure.Docker == nil && configure.Containerd == nil {
		node.Logger("the runtime container interface not found, config it use docker container .")
		configure.Docker = config.DefaultDockerConfiguration()
	}

	if configure.IsDockerCri() {
		docker.Install(configure.Docker, node, paths.China)
	} else {
		containerd.Install(configure.Docker, node, paths.China)
	}
}
