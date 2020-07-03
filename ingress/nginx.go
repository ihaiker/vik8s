package ingress

import (
	"fmt"
	"github.com/ihaiker/vik8s/install/repo"
	"github.com/ihaiker/vik8s/libs/flags"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/reduce"
	"github.com/spf13/cobra"
	"os"
)

type nginx struct {
	Repo    repo.Repo
	Version string `help:"" def:"0.30.0"`

	HostNetwork   bool `flag:"host-network" help:"deploy pod use hostNetwork"`
	NodePortHttp  int  `flag:"nodeport" help:"the ingress-nginx http 80 service nodeport, 0: automatic allocation, -1: disable" def:"-1"`
	NodePortHttps int  `flag:"nodeport-https" help:"the ingress-nginx https 443 service nodeport, 0: automatic allocation, -1: disable" def:"-1"`

	Replicas      int               `help:"ingress-nginx pod replicas number" def:"1"`
	NodeSelectors map[string]string `flag:"node.selector" help:"Deployment.nodeSelector"`
}

func (n *nginx) Name() string {
	return "nginx"
}

func (n *nginx) Description() string {
	return fmt.Sprintf("install kubernetes/ingress-nginx ( v0.30.0 ), more info see https://github.com/kubernetes/ingress-nginx")
}

func (n *nginx) Flags(cmd *cobra.Command) {
	flags.Flags(cmd.Flags(), n, "")
}

func (n *nginx) Apply(master *ssh.Node) {
	n.Repo.QuayIO("ingress-nginx")
	//name := "yaml/ingress/nginx.yaml"
	//tools.MustScpAndApplyAssert(master, name, n)
	name := "yaml/ingress/nginx.conf"
	reduce.MustApplyAssert(master, name, n)
}

func (n *nginx) Delete(master *ssh.Node) {
	err := master.CmdStd("kubectl delete namespaces ingress-nginx ", os.Stdout)
	utils.Panic(err, "delete nginx ingress")
}
