package pod

import (
	"github.com/ihaiker/vik8s/reduce/config"
	v1 "k8s.io/api/core/v1"
)

func HostAliasesParse(d *config.Directive, spec *v1.PodSpec) {
	for _, host := range d.Body {
		spec.HostAliases = append(spec.HostAliases, v1.HostAlias{
			IP: host.Name, Hostnames: host.Args,
		})
	}
}
