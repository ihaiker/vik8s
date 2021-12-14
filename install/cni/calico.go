package cni

import (
	"fmt"
	"github.com/ihaiker/cobrax"
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/install/paths"
	"github.com/ihaiker/vik8s/install/repo"
	"github.com/ihaiker/vik8s/install/tools"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/spf13/cobra"
)

const (
	tigera_operator  = "apply/cni/calico/tigera-operator.yaml"
	custom_resources = "apply/cni/calico/custom-resources.yaml"
)

type Calico struct {
	Version string `flag:"version" help:"calico server"`
	Repo    string `flag:"repo" help:"tigera/operator image repository. default: from quay.io or quay.mirrors.ustc.edu.cn in china"`
}

func NewCalico() *Calico {
	return &Calico{
		Version: "v1.23.1",
	}
}

func (f *Calico) Name() string {
	return "calico"
}

func (f *Calico) Flags(cmd *cobra.Command) {
	_ = cobrax.Flags(cmd, f, "", "")
}

func (f *Calico) Apply(configure *config.Configuration, node *ssh.Node) {
	image := fmt.Sprintf("%s/tigera/operator", repo.QuayIO(f.Repo))
	err := tools.ScpAndApplyAssert(node, "yaml/cni/calico/tigera-operator.yaml", paths.Json{
		"Image": image, "Version": f.Version,
	})
	utils.Panic(err, "apply calico error")

	err = tools.ScpAndApplyAssert(node, "yaml/cni/calico/custom-resources.yaml", paths.Json{
		"NetworkCidr": configure.K8S.PodCIDR,
	})
	utils.Panic(err, "apply calico custom resources error")
}

func (f *Calico) Clean(node *ssh.Node) {
	remote := node.Vik8s(tigera_operator)
	_ = node.CmdStdout("kubectl delete -f " + remote)

	remote = node.Vik8s(custom_resources)
	_ = node.CmdStdout("kubectl delete -f " + remote)
}
