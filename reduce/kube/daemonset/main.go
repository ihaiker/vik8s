package daemonset

import (
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/reduce/asserts"
	"github.com/ihaiker/vik8s/reduce/config"
	"github.com/ihaiker/vik8s/reduce/kube/pod"
	"github.com/ihaiker/vik8s/reduce/refs"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func podSpecParse(item *config.Directive, spec *appsv1.DaemonSetSpec) bool {
	return utils.Safe(func() { refs.Unmarshal(spec, item) }) == nil
}

func Parse(version, prefix string, directive *config.Directive) []metav1.Object {
	daemonset := &appsv1.DaemonSet{}
	asserts.Metadata(daemonset, directive)
	asserts.AutoLabels(daemonset, prefix)

	items := &config.Directive{}
	for {
		if d := directive.Body.Next(); d == nil {
			break
		} else {
			if !podSpecParse(d, &daemonset.Spec) {
				items.Body = append(items.Body, d)
			}
		}
	}
	services := pod.PodSpecParse(items, &daemonset.Spec.Template.Spec)

	daemonset.Spec.Template.Labels = daemonset.Labels
	daemonset.Spec.Template.Name = daemonset.Name
	daemonset.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: daemonset.Labels,
	}

	for _, object := range services {
		service := object.(*v1.Service)
		service.Labels = daemonset.Labels
		service.Spec.Selector = daemonset.Labels
	}

	return append([]metav1.Object{daemonset}, services...)
}
