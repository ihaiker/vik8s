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
	for _, item := range directive.Body {
		if handler, has := handlers[item.Name]; has {
			handler(item, podSpec)
		} else {
			refs.UnmarshalItem(podSpec, item)
		}
	}
}

func Parse(version, prefix string, directive *config.Directive) []metav1.Object {
	pod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: v1.SchemeGroupVersion.String(),
		},
	}
	asserts.Metadata(pod, directive)
	asserts.AutoLabels(pod, prefix)
	PodSpecParse(directive, &pod.Spec)
	return []metav1.Object{pod}
}
