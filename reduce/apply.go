package reduce

import (
	"github.com/ihaiker/vik8s/install/tools"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/reduce/kube"
	"os"
)

func MustApplyAssert(node *ssh.Node, name string, data interface{}) {
	pods := tools.MustAssert(name, data)
	remote := node.Vik8s("apply", name[5:len(name)-5]+".yaml")

	pods = kube.ParseWith(pods).Bytes()
	err := node.ScpContent(pods, remote)
	utils.Panic(err, "scp %s", name)

	err = node.CmdOutput("kubectl apply -f "+remote, os.Stdout)
	utils.Panic(err, "kubectl apply -f %s", remote)
}
