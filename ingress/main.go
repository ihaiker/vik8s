package ingress

import (
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/spf13/cobra"
)

type Ingress interface {
	Name() string
	Description() string
	Flags(cmd *cobra.Command)
	Apply(master *ssh.Node)
	Delete(master *ssh.Node)
}

type manager []Ingress

var Manager = manager{
	new(traefik), new(nginx),
}

func (p *manager) Apply(name string, master *ssh.Node) {
	for _, plugin := range *p {
		if plugin.Name() == name {
			plugin.Apply(master)
			return
		}
	}
}

func (p *manager) Delete(name string, master *ssh.Node) {
	for _, plugin := range *p {
		if plugin.Name() == name {
			plugin.Delete(master)
			return
		}
	}
}
