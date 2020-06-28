package pod

import (
	"github.com/ihaiker/vik8s/reduce/asserts"
	"github.com/ihaiker/vik8s/reduce/config"
	"github.com/ihaiker/vik8s/reduce/kube/pod/container"
	"github.com/ihaiker/vik8s/reduce/kube/pod/volumes"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var handlers = map[string]func(directive *config.Directive, pod *v1.Pod){
	"container":      container.ContainerParse,
	"init-container": container.InitContainerParse, "initContainer": container.InitContainerParse,

	"nodeSelector": nodeSelectorParse,

	"hosts":  HostAliasesParse,
	"volume": volumes.VolumeParse, "volumes": volumes.VolumesParse,

	"affinity": AffinityParse,
}

func Parse(directive *config.Directive) metav1.Object {
	pod := &v1.Pod{}
	asserts.Metadata(pod, directive)
	for _, d := range directive.Body {
		if handler, has := handlers[d.Name]; has {
			handler(d, pod)
		}
	}
	return pod
}
