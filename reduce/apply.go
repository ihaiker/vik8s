package reduce

import (
	"github.com/ihaiker/vik8s/install/tools"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/reduce/kube"
	"os"
	"strings"
	"text/template"
)

func MustApplyAssert(node *ssh.Node, name string, data interface{}, funcs ...template.FuncMap) {
	pods := tools.MustAssert(name, data, funcs...)
	remote := node.Vik8s("apply", strings.TrimPrefix(name+".yaml", "yaml/"))

	pods = kube.ParseWith(pods).Bytes()
	err := node.ScpContent(pods, remote)
	utils.Panic(err, "scp %s", name)

	err = node.CmdStd("kubectl apply -f "+remote, os.Stdout)
	utils.Panic(err, "kubectl apply -f %s", remote)
}
