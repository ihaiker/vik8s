package kube

import (
	"github.com/ihaiker/vik8s/reduce/asserts"
	"github.com/ihaiker/vik8s/reduce/config"
)

type Node struct {
	*Entry
}

func (c *Node) ToYaml(indent int) string {
	w := config.Writer(indent)
	w.Writer(c.Entry.Yaml(indent))
	return w.String()
}

func nodeParse(d *config.Directive, kube *Kubernetes) {
	asserts.ArgsMin(d, 1)
	node := &Node{
		Entry: &Entry{Name: d.Args[0],
			Labels: Labels(), Annotations: Annotations(),
		},
	}
	bodyLabels(node.Labels, d)
	entry(d, node.Entry, nil)
	kube.Add(node)
}

func init() {
	kinds["node"] = nodeParse
	kinds["node"] = nodeParse
}
