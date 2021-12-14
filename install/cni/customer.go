package cni

import (
	"fmt"
	"github.com/ihaiker/cobrax"
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/spf13/cobra"
	"net/url"
	"path/filepath"
)

type Customer struct {
	Urls  []string `flag:"url" help:"User-defined network plugin URL, if used for kubectl apply -f <url>"`
	Files []string `flag:"file" help:"User-defined network plugin file location, if used for kubectl apply -f <file>"`
}

func (f *Customer) Name() string {
	return "customer"
}

func (f *Customer) Flags(cmd *cobra.Command) {
	err := cobrax.Flags(cmd, f, "", "")
	utils.Panic(err, "set customer flag")
}

func (f *Customer) Apply(configure *config.Configuration, node *ssh.Node) {
	utils.Assert(len(f.Files) != 0 || len(f.Urls) != 0,
		"No custom network plugin found，see: --url or --file ")

	remote := node.Vik8s("yaml/cni/customer")
	err := node.HideLog().Cmd(fmt.Sprintf("rm -rf %s | mkdir -p %s", remote, remote))
	utils.Panic(err, "mkdir customer network config file directory")

	for _, file := range f.Files {
		name := filepath.Base(file)
		err := node.Scp(file, filepath.Join(remote, name))
		utils.Panic(err, "apply customer network")
	}

	for _, urlString := range f.Urls {
		u, err := url.Parse(urlString)
		utils.Panic(err, "--url %s", urlString)
		name := filepath.Base(u.Path)
		err = node.Cmd(fmt.Sprintf("curl -o %s %s", filepath.Join(remote, name), urlString))
	}

	err = node.CmdStdout(fmt.Sprintf("kubectl apply -k %s", remote))
	utils.Panic(err, "apply customer network")
}

func (f *Customer) Clean(node *ssh.Node) {
	remote := node.Vik8s("yaml/cni/customer")
	_ = node.CmdStdout(fmt.Sprintf("kubectl delete -k %s", remote))
}
