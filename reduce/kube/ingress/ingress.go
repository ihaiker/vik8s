package ingress

import (
	"github.com/ihaiker/vik8s/reduce/asserts"
	"github.com/ihaiker/vik8s/reduce/config"
	"k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func IngressParse(version, prefix string, dir *config.Directive) []metav1.Object {
	i := &v1beta1.Ingress{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Ingress",
			APIVersion: v1beta1.SchemeGroupVersion.String(),
		},
	}
	asserts.Metadata(i, dir)
	asserts.AutoLabels(i, prefix)
	return []metav1.Object{i}
}
