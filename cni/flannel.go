package cni

import (
	"github.com/ihaiker/vik8s/install/k8s"
	"github.com/ihaiker/vik8s/install/repo"
	"github.com/ihaiker/vik8s/install/tools"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/spf13/cobra"
)

type flannel struct {
	version     string
	repo        string
	limitCPU    string
	limitMemory string
}

func (f *flannel) Name() string {
	return "flannel"
}

func (f *flannel) Flags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.version, flags(f, "version"), "0.12.0", "the flannel version")
	cmd.Flags().StringVar(&f.limitCPU, flags(f, "limits-cpu"), "100m", "Container Cup Limit")
	cmd.Flags().StringVar(&f.limitMemory, flags(f, "limits-memory"), "50Mi", "Container Memory Limit")
	cmd.Flags().StringVar(&f.repo, flags(f, "repo"), "", "")
}

func (f *flannel) Apply(node *ssh.Node) {

	data := tools.Json{
		"Version": f.version, "Repo": repo.QuayIO(f.repo),
		"CIDR": k8s.Config.Kubernetes.PodCIDR, "Interface": k8s.Config.Kubernetes.Interface,
		"LimitCPU": f.limitCPU, "LimitMemory": f.limitMemory,
	}

	name := "yaml/cni/flannel.yaml"
	tools.MustScpAndApplyAssert(node, name, data)

	k8s.Config.CNI.Params = map[string]string{
		"Version": f.version, "Repo": repo.QuayIO(f.repo),
		"LimitCPU": f.limitCPU, "LimitMemory": f.limitMemory,
	}
}

func (f *flannel) Clean(node *ssh.Node) {
	_, _ = node.Cmd("ifconfig flannel.1 down")
	_, _ = node.Cmd("ip link delete flannel.1")
	_, _ = node.Cmd("rm -rf /var/lib/cni/ /etc/cni/net.d/*")
}
