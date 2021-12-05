package k8s

import (
	"fmt"
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
)

func RemoveTaint(configure *config.Configuration, node *ssh.Node) {
	master := configure.Hosts.MustGet(configure.K8S.Masters[0])
	err := master.CmdStdout(fmt.Sprintf("kubectl taint node %s node-role.kubernetes.io/master-", node.Hostname))
	utils.Panic(err, "remove taint")
}
