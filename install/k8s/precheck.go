package k8s

import (
	"github.com/ihaiker/vik8s/install/bases"
	"github.com/ihaiker/vik8s/install/cri"
	"github.com/ihaiker/vik8s/libs/ssh"
)

func preCheck(node *ssh.Node) {
	bases.Check(node)
	cri.Install(node)
	checkKubernetes(node)
}
