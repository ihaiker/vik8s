package pod

import (
	"github.com/ihaiker/vik8s/reduce/config"
	v1 "k8s.io/api/core/v1"
)

func HostAliasesParse(d *config.Directive, pod *v1.Pod) {
	for _, host := range d.Body {
		pod.Spec.HostAliases = append(pod.Spec.HostAliases, v1.HostAlias{
			IP: host.Name, Hostnames: host.Args,
		})
	}
}
