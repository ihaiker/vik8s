package kube

import (
	"github.com/ihaiker/vik8s/reduce/asserts"
	"github.com/ihaiker/vik8s/reduce/config"
)

type Namespace struct {
	*Entry
}

func (n *Namespace) ToYaml(indent int) string {
	return n.Entry.Yaml(indent)
}

type Yaml struct {
	*Entry
	Body string
}

func (n *Yaml) ToYaml(indent int) string {
	return n.Body[1 : len(n.Body)-1]
}

func namespaceParse(d *config.Directive, kube *Kubernetes) {
	asserts.ArgsMin(d, 1)
	namespace := &Namespace{
		Entry: &Entry{
			Name:        d.Args[0],
			Labels:      Labels(),
			Annotations: NewProperties(""),
		},
	}
	argsLabels(namespace.Labels, d.Args[1:])
	entry(d, namespace.Entry, func(body *config.Directive) {
		asserts.ArgsLen(body, 1)
		namespace.Labels.Add(body.Name, body.Args[0])
	})

	kube.Objects = append(kube.Objects, namespace)
}

func init() {
	kinds["namespace"] = namespaceParse
	kinds["Namespace"] = namespaceParse
	kinds["yaml"] = func(d *config.Directive, kube *Kubernetes) {
		yaml := &Yaml{
			Entry: &Entry{}, Body: "",
		}
		yaml.Body = d.Args[0]
		kube.Objects = append(kube.Objects, yaml)
	}
}
