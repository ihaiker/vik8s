package k8s

import (
	"github.com/ihaiker/vik8s/install"
	"github.com/ihaiker/vik8s/libs/ssh"
)

func preCheck(node *ssh.Node) {
	install.PreCheck(node)
	checkDocker(node)
	checkKubernetes(node)
}
