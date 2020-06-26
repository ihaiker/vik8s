package kube

import (
	"github.com/ihaiker/vik8s/reduce/asserts"
	"github.com/ihaiker/vik8s/reduce/config"
)

type (
	YamlEntry interface {
		ToYaml(indent int) string
	}

	IEntry interface {
		Set(namespace, kind, version string)
	}

	Entry struct {
		Name          string
		Kind, Version string

		Namespace   string
		Labels      *Properties
		Annotations *Properties
	}
)

func (entry *Entry) Set(namespace, kind, version string) {
	entry.Namespace = namespace
	entry.Version = version
	entry.Kind = kind
}

func (entry *Entry) Yaml(indent int) string {
	w := config.Writer(indent)
	w.Line("apiVersion:", entry.Version)
	w.Line("kind:", entry.Kind)
	w.Line("metadata:")
	w.Tab().Line("name:", entry.Name)
	if entry.Kind != "Namespace" && entry.Namespace != "" {
		w.Tab().Line("namespace:", entry.Namespace)
	}
	w.Writer(entry.Labels.ToYaml(1))
	w.Writer(entry.Annotations.ToYaml(1))
	return w.String()
}

func entry(d *config.Directive, entry *Entry, bodyfn func(body *config.Directive)) {
	for _, body := range d.Body {
		switch body.Name {
		case "namespace":
			asserts.ArgsLen(body, 1)
			entry.Namespace = body.Args[0]
		case "annotations":
			annotations(entry.Annotations, body)
		case "annotation":
			asserts.ArgsLen(body, 2)
			entry.Annotations.Add(body.Args[0], body.Args[1])
		case "labels":
			bodyLabels(entry.Labels, body)
		case "label":
			asserts.ArgsLen(body, 2)
			entry.Labels.Add(body.Args[0], body.Args[1])
		default:
			if bodyfn != nil {
				bodyfn(body)
			}
		}
	}
}
