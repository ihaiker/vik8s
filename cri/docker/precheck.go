package docker

import (
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/libs/ssh"
)

//https://docs.docker.com/config/daemon/

//Install docker server to node, cfg is configuration.
func Install(cfg *config.DockerConfiguration, node *ssh.Node, china bool) error {

	return nil
}
