package cri

import (
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/install/cri/containerd"
	"github.com/ihaiker/vik8s/install/cri/docker"
	"github.com/ihaiker/vik8s/install/paths"
	"github.com/ihaiker/vik8s/libs/ssh"
)

func Install(node *ssh.Node) {
	if config.Config.Docker == nil && config.Config.Containerd == nil {
		node.Logger("the runtime container interface not found, config it use docker container .")
		config.Config.Docker = config.DefaultDockerConfiguration()
	}

	if config.Config.IsDockerCri() {
		docker.Install(config.Config.Docker, node, paths.China)
	} else {
		containerd.Install(config.Config.Docker, node, paths.China)
	}
}
