package ingress

import (
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/reduce/asserts"
	"github.com/ihaiker/vik8s/reduce/config"
	"github.com/ihaiker/vik8s/reduce/refs"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

func Parse(version, prefix string, directive *config.Directive) []metav1.Object {
	ig := &networkingv1beta1.Ingress{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Ingress",
			APIVersion: networkingv1beta1.SchemeGroupVersion.String(),
		},
	}
	asserts.Metadata(ig, directive)
	asserts.AutoLabels(ig, prefix)

	for _, d := range directive.Body {
		switch d.Name {
		case "tls":
			if d.HasArgs() {
				ig.Spec.TLS = append(ig.Spec.TLS, networkingv1beta1.IngressTLS{
					Hosts: d.Args[1:], SecretName: d.Args[0],
				})
			} else {
				for _, sub := range d.Body {
					ig.Spec.TLS = append(ig.Spec.TLS, networkingv1beta1.IngressTLS{
						SecretName: sub.Name, Hosts: sub.Args,
					})
				}
			}
		case "rules":
			asserts.ArgsLen(d, 1)
			rule := networkingv1beta1.IngressRule{
				Host: d.Args[0], IngressRuleValue: networkingv1beta1.IngressRuleValue{
					HTTP: &networkingv1beta1.HTTPIngressRuleValue{},
				},
			}
			for _, path := range d.Body {
				utils.Assert(path.Name == "http" && len(path.Args) >= 1 && path.Args[0] == "paths",
					"Invalid parameter: %s %s , line %d",
					path.Name, strings.Join(path.Args, " "), path.Line)
				ingressPath := networkingv1beta1.HTTPIngressPath{Path: utils.Index(path.Args, 1)}
				refs.Unmarshal(&ingressPath.Backend, path)
				rule.IngressRuleValue.HTTP.Paths = append(rule.IngressRuleValue.HTTP.Paths, ingressPath)
			}
			ig.Spec.Rules = append(ig.Spec.Rules, rule)
		case "backend":
			ig.Spec.Backend = &networkingv1beta1.IngressBackend{}
			refs.Unmarshal(ig.Spec.Backend, d)
		}
	}
	return []metav1.Object{ig}
}
