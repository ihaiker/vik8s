package kube

import (
	"github.com/ihaiker/vik8s/reduce/asserts"
	"github.com/ihaiker/vik8s/reduce/config"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Node struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
}

func nodeParse(version, prefix string, directive *config.Directive) []metav1.Object {
	node := &Node{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Node",
			APIVersion: v1.SchemeGroupVersion.String(),
		},
	}
	asserts.Metadata(node, directive)
	for _, d := range directive.Body {
		node.Labels[d.Name] = d.Args[0]
	}
	return []metav1.Object{node}
}
