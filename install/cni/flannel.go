package cni

import (
	"github.com/ihaiker/cobrax"
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/install/paths"
	"github.com/ihaiker/vik8s/install/repo"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/reduce"
	"github.com/spf13/cobra"
)

type flannel struct {
	Version     string `flag:"version" help:"the flannel version"`
	Repo        string `flag:"repo" help:"docker image pull from."`
	LimitCPU    string `flag:"limits-cpu" help:"Container Cup Limit"`
	LimitMemory string `flag:"limits-memory" help:"Container Memory Limit"`
}

func NewFlannelCni() *flannel {
	return &flannel{
		Version:     "v0.14.0",
		LimitCPU:    "100m",
		LimitMemory: "50Mi",
	}
}
func (f *flannel) Name() string {
	return "flannel"
}

func (f *flannel) Flags(cmd *cobra.Command) {
	err := cobrax.Flags(cmd, f, "", "")
	utils.Panic(err, "set flannel flag error")
}

func (f *flannel) Apply(cmd *cobra.Command, configure *config.Configuration, node *ssh.Node) {
	data := paths.Json{
		"Version": f.Version, "Repo": repo.QuayIO(f.Repo),
		"CIDR": configure.K8S.PodCIDR, "Interface": configure.K8S.Interface,
		"LimitCPU": f.LimitCPU, "LimitMemory": f.LimitMemory,
	}
	name := "yaml/cni/flannel.conf"
	err := reduce.ApplyAssert(node, name, data)
	utils.Panic(err, "apply flannel network")
}

func (f *flannel) Clean(node *ssh.Node) {
	_ = node.Sudo().CmdStdout("ifconfig flannel.1 down")
	_ = node.Sudo().CmdStdout("ip link delete flannel.1")
	_ = node.Sudo().CmdStdout("rm -rf /var/lib/cni/ /etc/cni/net.d/*")
}
