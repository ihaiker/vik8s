package kube

import (
	"github.com/ihaiker/vik8s/reduce/asserts"
	"github.com/ihaiker/vik8s/reduce/config"
	"github.com/ihaiker/vik8s/reduce/plugins"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var Namespace = plugins.ReduceHandler{
	Names: []string{"namespace", "Namespace"},
	Handler: func(version, prefix string, directive *config.Directive) metav1.Object {
		asserts.ArgsMin(directive, 1)
		namespace := &v1.Namespace{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Namespace",
				APIVersion: v1.SchemeGroupVersion.String(),
			},
		}
		asserts.Metadata(namespace.GetObjectMeta(), directive)

		labels := namespace.GetLabels()
		for _, d := range directive.Body {
			labels[d.Name] = d.Args[0]
		}
		namespace.SetLabels(labels)
		return namespace
	},
	Demo: `
namespace name;
namespace name [label1=value1 label2=label3 ...] {
	label-sub-1 value-sub-1;
	label-sub-2 value-sub-2;
	...
}
`,
}
