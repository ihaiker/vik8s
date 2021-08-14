package cni

import (
	"fmt"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/spf13/cobra"
)

type customer struct {
	url  string
	file string
}

func (f *customer) Name() string {
	return "customer"
}

func (f *customer) Flags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.url, flags(f, "url"), "", "User-defined network plugin URL, if used for kubectl apply -f <url>")
	cmd.Flags().StringVar(&f.file, flags(f, "file"), "", "User-defined network plugin file location, if used for kubectl apply -f <file>")
}

func (f *customer) Apply(cmd *cobra.Command, node *ssh.Node) {
	utils.Assert(f.url != "" || f.file != "", "No custom network plugin found，see: --customer-url or --customer-file ")

	remote := node.Vik8s("yaml/cni/customer.yaml")
	if f.url != "" {
		remote = f.url
	} else if f.file != "" {
		err := node.Scp(f.file, remote)
		utils.Panic(err, "apply customer network")
	}
	err := node.CmdStdout(fmt.Sprintf("kubectl apply -f %s", remote))
	utils.Panic(err, "apply customer network")
}

func (f *customer) Clean(node *ssh.Node) {}
