package reduce

import (
	"github.com/ihaiker/vik8s/install/tools"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/reduce/kube"
)

func ApplyAssert(node *ssh.Node, name string, data interface{}) error {
	pods, err := tools.Assert(name, data)
	if err != nil {
		return err
	}

	remote := node.Vik8s("apply", name[5:len(name)-5]+".yaml")
	pods = kube.ParseWith(pods).Bytes()
	if err = node.HideLog().ScpContent(pods, remote); err != nil {
		return err
	}
	err = node.CmdStdout("kubectl apply -f " + remote)
	return err
}
