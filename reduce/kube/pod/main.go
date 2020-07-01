package pod

import (
	"github.com/ihaiker/vik8s/reduce/asserts"
	"github.com/ihaiker/vik8s/reduce/config"
	"github.com/ihaiker/vik8s/reduce/kube/pod/container"
	"github.com/ihaiker/vik8s/reduce/kube/pod/volumes"
	"github.com/ihaiker/vik8s/reduce/refs"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var handlers = map[string]func(directive *config.Directive, spec *v1.PodSpec){
	"container": container.ContainerParse,

	"init-container": container.InitContainerParse, "initContainer": container.InitContainerParse,

	"hosts":  HostAliasesParse,
	"volume": volumes.VolumeParse, "volumes": volumes.VolumesParse,
	"affinity": AffinityParse,
}

func PodSpecParse(directive *config.Directive, podSpec *v1.PodSpec) {
	for _, d := range directive.Body {
		if handler, has := handlers[d.Name]; has {
			handler(d, podSpec)
		} else {
			refs.Unmarshal(podSpec, d)
		}
	}
}

func Parse(directive *config.Directive) metav1.Object {
	pod := &v1.Pod{}
	asserts.Metadata(pod, directive)
	PodSpecParse(directive, &pod.Spec)
	return pod
}
