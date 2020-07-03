package kube

import (
	"github.com/ihaiker/vik8s/reduce/asserts"
	"github.com/ihaiker/vik8s/reduce/config"
	"github.com/ihaiker/vik8s/reduce/plugins"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type node struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
}

var Node = plugins.ReduceHandler{
	Names: []string{"node", "Node"}, Handler: func(version, prefix string, directive *config.Directive) metav1.Object {
		node := &node{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Node",
				APIVersion: v1.SchemeGroupVersion.String(),
			},
		}
		asserts.Metadata(node, directive)
		for _, d := range directive.Body {
			node.Labels[d.Name] = d.Args[0]
		}
		return node
	},
	Demo: `
node nodeName [label1=value1 ...]{
	lable-l1 value-v1;
	lable-l2 value-v2;
}
`,
}
