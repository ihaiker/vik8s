package kube

import (
	"github.com/ihaiker/vik8s/reduce/asserts"
	"github.com/ihaiker/vik8s/reduce/config"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func namespaceParse(version, prefix string, directive *config.Directive) []metav1.Object {
	asserts.ArgsMin(directive, 1)
	namespace := &v1.Namespace{}
	asserts.Metadata(namespace.GetObjectMeta(), directive)

	labels := namespace.GetLabels()
	for _, d := range directive.Body {
		labels[d.Name] = d.Args[0]
	}
	namespace.SetLabels(labels)
	return []metav1.Object{namespace}
}
