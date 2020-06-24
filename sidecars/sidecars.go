package sidecars

import (
	"fmt"
	"github.com/ihaiker/vik8s/sidecars/dashboard"
	"github.com/spf13/cobra"
)

type Sidecars interface {
	Name() string
	Description() string
	Flags(cmd *cobra.Command)
	Apply()
	Delete(data bool)
}

type sidecars []Sidecars

var Manager = sidecars{
	new(dashboard.Dashboard),
}

func (p *sidecars) Apply(name string) {
	fmt.Println("apply ", name)
	for _, plugin := range *p {
		if plugin.Name() == name {
			plugin.Apply()
			return
		}
	}
}

func (p *sidecars) Delete(name string, data bool) {
	fmt.Println("delete ", name)
	for _, plugin := range *p {
		if plugin.Name() == name {
			plugin.Delete(data)
			return
		}
	}
}
