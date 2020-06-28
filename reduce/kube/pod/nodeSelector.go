package pod

import (
	"github.com/ihaiker/vik8s/reduce/config"
	v1 "k8s.io/api/core/v1"
)

func nodeSelectorParse(d *config.Directive, pod *v1.Pod) {
	if pod.Spec.NodeSelector == nil {
		pod.Spec.NodeSelector = make(map[string]string)
	}

	for _, directive := range d.Body {
		pod.Spec.NodeSelector[directive.Name] = directive.Args[0]
	}
}
