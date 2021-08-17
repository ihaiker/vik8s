package cni

import (
	"fmt"
	"github.com/ihaiker/cobrax"
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/install/repo"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/spf13/cobra"
	"path/filepath"
)

const (
	tigera_operator  = "apply/cni/calico/tigera-operator.yaml"
	custom_resources = "apply/cni/calico/custom-resources.yaml"
)

type calico struct {
	OperatorVersion string `flag:"operator-version" help:"calico operator server version"`
	Repo            string `flag:"repo" help:"tigera/operator image repository. default: from quay.io or quay.mirrors.ustc.edu.cn in china"`
	Version         string `flag:"version" help:"calico server"`
}

func NewCalico() Plugin {
	return &calico{
		OperatorVersion: "v1.20.0",
	}
}

func (f *calico) Name() string {
	return "calico"
}

func (f *calico) Flags(cmd *cobra.Command) {
	_ = cobrax.Flags(cmd, f, "", "")
}

func (f *calico) Apply(cmd *cobra.Command, node *ssh.Node) {
	remote := node.Vik8s(tigera_operator)
	operatorUrl := "https://docs.projectcalico.org/manifests/tigera-operator.yaml"

	node.Logger("download calico tigera operator config yaml %s", operatorUrl)
	err := node.Cmd(fmt.Sprintf("mkdir -p %s | curl -o %s %s", filepath.Dir(remote), remote, operatorUrl))
	utils.Panic(err, "get operator config error")

	image := fmt.Sprintf("%s/tigera/operator:%s", repo.QuayIO(f.Repo), f.OperatorVersion)
	node.Logger("modify config image version: %s", image)
	selectPattern := `(select(.kind == "Deployment" and .metadata.name == "tigera-operator") | .spec.template.spec.containers[0].image)`
	err = node.Cmd(fmt.Sprintf(`yq e -i '%s = "%s"' %s`, selectPattern, image, remote))
	utils.Panic(err, "modify %s", operatorUrl)

	node.Logger("apply calico network interface # tigera operator ")
	err = node.CmdStdout("kubectl apply -f " + remote)
	utils.Panic(err, "kubectl apply error")

	//---
	remote = node.Vik8s(custom_resources)
	operatorUrl = "https://docs.projectcalico.org/manifests/custom-resources.yaml"
	node.Logger("download calico custom resources config yaml %s", operatorUrl)
	err = node.Cmd(fmt.Sprintf("mkdir -p %s | curl -o %s %s", filepath.Dir(remote), remote, operatorUrl))
	utils.Panic(err, "download calico custom resource config error")

	err = node.Cmd(fmt.Sprintf(`yq e -i '(select(.kind == "Installation") | .spec.calicoNetwork.ipPools[0].cidr) = "%s"' %s`,
		config.K8S().PodCIDR, remote))
	utils.Panic(err, "modify custom resource error")

	node.Logger("apply calico network interface # resource ")
	err = node.CmdStdout("kubectl apply -f " + remote)
	utils.Panic(err, "kubectl apply error")
}

func (f *calico) Clean(node *ssh.Node) {
	remote := node.Vik8s(tigera_operator)
	_ = node.CmdStdout("kubectl delete -f " + remote)

	remote = node.Vik8s(custom_resources)
	_ = node.CmdStdout("kubectl delete -f " + remote)
}
