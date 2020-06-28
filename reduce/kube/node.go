package kube

import (
	"github.com/ihaiker/vik8s/reduce/asserts"
	"github.com/ihaiker/vik8s/reduce/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Node struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
}

func nodeParse(directive *config.Directive) metav1.Object {
	node := &Node{}
	asserts.Metadata(node, directive)
	for _, d := range directive.Body {
		node.Labels[d.Name] = d.Args[0]
	}
	return node
}
