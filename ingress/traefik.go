package ingress

import (
	"fmt"
	"github.com/ihaiker/vik8s/install/repo"
	"github.com/ihaiker/vik8s/install/tools"
	"github.com/ihaiker/vik8s/libs/flags"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/spf13/cobra"
	"math/rand"
	"os"
	"strings"
)

type traefik struct {
	Repo          repo.Repo
	Version       string            `def:"1.7"`
	HostNetwork   bool              `flag:"host-network" help:"Whether to enable the host network method"`
	NodePortHttp  int               `flag:"nodeport" help:"the ingress-traefik http service nodeport, 0: automatic allocatiot,  -1: disable" def:"-1"`
	NodePortHttps int               `flag:"nodeport-https" help:"the ingress-traefik https 443 service nodeport, 0: automatic allocation, -1: disable" def:"-1"`
	Replicas      int               `help:"ingress-traefik pod replicas number" def:"1"`
	NodeSelectors map[string]string `flag:"node.selector"`

	IngressUI    string `flag:"ui-ingress" help:"Creating ingress that will expose the Traefik Web UI."`
	AuthUI       bool   `flag:"ui-auth" help:"Whether to enable 'basic authentication' in traefik web ui ingress" def:"true"`
	AuthUser     string `flag:"ui-user" help:"web ui 'basic authentication' user " def:"admin"`
	AuthPassword string `flag:"ui-passwd" help:"web ui 'basic authentication' password (default: randomly generated and pint to console)"`
}

func (t *traefik) Name() string {
	return "traefik"
}

func (t *traefik) Description() string {
	return "https://docs.traefik.io/v1.7/user-guide/kubernetes/"
}

func (t *traefik) Flags(cmd *cobra.Command) {
	flags.Flags(cmd.Flags(), t, "")
}

func (t *traefik) Apply(master *ssh.Node) {
	t.Repo.Set("ingress-traefik")

	if t.AuthUI {
		if t.AuthPassword == "" {
			t.AuthPassword = fmt.Sprintf("%06d", rand.Int63n(1000000))
			fmt.Println(strings.Repeat("=", 40))
			fmt.Printf("  the web ui ingress default password for `%s` is : %s\n", t.AuthUser, t.AuthPassword)
			fmt.Println(strings.Repeat("=", 40))
		}

		t.AuthPassword, _ = utils.HashApr1(t.AuthPassword)
	}

	name := "yaml/ingress/traefik.yaml"
	tools.MustScpAndApplyAssert(master, name, t)
}

func (n *traefik) Delete(master *ssh.Node) {
	err := master.CmdStd("kubectl delete namespaces ingress-traefik ", os.Stdout)
	utils.Panic(err, "delete traefik ingress")
}
