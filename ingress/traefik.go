package ingress

import (
	"fmt"
	"github.com/ihaiker/cobrax"
	"github.com/ihaiker/vik8s/install/repo"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/reduce"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

type traefik struct {
	Repo          repo.Repo
	Version       string            `help:"traefik ingress version"`
	HostNetwork   bool              `flag:"host-network" help:"Whether to enable the host network method"`
	NodePortHttp  int               `flag:"nodeport" help:"the ingress-traefik http service nodeport, 0: automatic allocatiot,  -1: disable" def:"-1"`
	NodePortHttps int               `flag:"nodeport-https" help:"the ingress-traefik https 443 service nodeport, 0: automatic allocation, -1: disable" def:"-1"`
	Replicas      int               `help:"ingress-traefik pod replicas number" def:"1"`
	NodeSelectors map[string]string `flag:"node-selector" help:"select what node to deploy"`

	IngressUI    string `flag:"ui-ingress" help:"Creating ingress that will expose the Traefik Web UI."`
	AuthUI       bool   `flag:"ui-auth" help:"Whether to enable 'basic authentication' in traefik web ui ingress"`
	AuthUser     string `flag:"ui-user" help:"web ui 'basic authentication' user "`
	AuthPassword string `flag:"ui-passwd" help:"web ui 'basic authentication' password (default: randomly generated and pint to console)"`
}

func Treafik() Ingress {
	return &traefik{
		Version:     "v1.7",
		HostNetwork: false, NodePortHttp: -1, NodePortHttps: -1,
		Replicas: 1, NodeSelectors: map[string]string{},
		AuthUI: true, IngressUI: "traefik.vik8s.io", AuthUser: "admin",
	}
}

func (t *traefik) Name() string {
	return "traefik"
}

func (t *traefik) Description() string {
	return "https://docs.traefik.io/v1.7/user-guide/kubernetes/"
}

func (t *traefik) Flags(cmd *cobra.Command) {
	err := cobrax.Flags(cmd, t, "", "")
	utils.Panic(err, "set treafik ingress error")
}

func (t *traefik) Apply(master *ssh.Node) {
	t.Repo.Set("ingress-traefik")

	if t.AuthUI {
		if t.AuthPassword == "" {
			t.AuthPassword = utils.Random(8)
			fmt.Println(strings.Repeat("=", 40))
			fmt.Printf("the web ui ingress default password for `%s` is : %s\n", t.AuthUser, t.AuthPassword)
			fmt.Println(strings.Repeat("=", 40))
		}
		t.AuthPassword, _ = utils.HashApr1(t.AuthPassword)
	}

	//name := "yaml/ingress/traefik.yaml"
	//tools.MustScpAndApplyAssert(master, name, t)
	name := "yaml/ingress/traefik.conf"
	err := reduce.ApplyAssert(master, name, t)
	utils.Panic(err, "apply traefik ingress")
}

func (n *traefik) Delete(master *ssh.Node) {
	err := master.CmdOutput("kubectl delete namespaces ingress-traefik ", os.Stdout)
	utils.Panic(err, "delete traefik ingress")
}
