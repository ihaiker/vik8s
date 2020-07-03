package daemonset

import (
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/reduce/asserts"
	"github.com/ihaiker/vik8s/reduce/config"
	"github.com/ihaiker/vik8s/reduce/kube/pod"
	"github.com/ihaiker/vik8s/reduce/refs"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func specParse(item *config.Directive, spec *appsv1.DaemonSetSpec) bool {
	return utils.Safe(func() { refs.UnmarshalItem(spec, item) }) == nil
}

func Parse(version, prefix string, directive *config.Directive) []metav1.Object {
	daemonset := &appsv1.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: appsv1.SchemeGroupVersion.String(),
		},
	}
	asserts.Metadata(daemonset, directive)
	asserts.AutoLabels(daemonset, prefix)

	for it := directive.Body.Iterator(); it.HasNext(); {
		d := it.Next()
		if specParse(d, &daemonset.Spec) {
			it.Remove()
		}
	}
	pod.PodSpecParse(directive, &daemonset.Spec.Template.Spec)

	daemonset.Spec.Template.Labels = daemonset.Labels
	daemonset.Spec.Template.Name = daemonset.Name
	daemonset.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: daemonset.Labels,
	}
	return []metav1.Object{daemonset}
}
