package deployment

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

func specParse(item *config.Directive, spec *appsv1.DeploymentSpec) bool {
	return utils.Safe(func() { refs.UnmarshalItem(spec, item) }) == nil
}

func Parse(version, prefix string, directive *config.Directive) []metav1.Object {
	dep := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: appsv1.SchemeGroupVersion.String(),
		},
	}

	asserts.Metadata(dep, directive)
	asserts.AutoLabels(dep, prefix)

	for it := directive.Body.Iterator(); it.HasNext(); {
		d := it.Next()
		if specParse(d, &dep.Spec) {
			it.Remove()
		}
	}
	services := pod.PodSpecParse(directive, &dep.Spec.Template.Spec)

	dep.Spec.Template.Labels = dep.Labels
	dep.Spec.Template.Name = dep.Name
	dep.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: dep.Labels,
	}

	for _, object := range services {
		service := object.(*v1.Service)
		service.Labels = dep.Labels
		service.Spec.Selector = dep.Labels
	}

	return append([]metav1.Object{dep}, services...)
}
