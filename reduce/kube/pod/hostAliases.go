package pod

import (
	"github.com/ihaiker/vik8s/reduce/config"
	v1 "k8s.io/api/core/v1"
)

func HostAliasesParse(d *config.Directive, spec *v1.PodSpec) {
	if len(d.Args) > 0 {
		spec.HostAliases = append(spec.HostAliases, v1.HostAlias{
			IP: d.Args[0], Hostnames: d.Args[1:],
		})
	}
	for _, host := range d.Body {
		spec.HostAliases = append(spec.HostAliases, v1.HostAlias{
			IP: host.Name, Hostnames: host.Args,
		})
	}
}
