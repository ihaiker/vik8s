package kube

import (
	"github.com/ihaiker/vik8s/reduce/asserts"
	"github.com/ihaiker/vik8s/reduce/config"
)

type ConfigMap struct {
	*Entry
	Data Data
}

func (c *ConfigMap) ToYaml(indent int) string {
	w := config.Writer(indent)
	w.Writer(c.Entry.Yaml(indent))
	w.Writer(c.Data.ToYaml(indent))
	return w.String()
}

func configmapParse(d *config.Directive, kube *Kubernetes) {
	asserts.ArgsMin(d, 1)
	configmap := &ConfigMap{
		Entry: &Entry{
			Name:        d.Args[0],
			Labels:      Labels(),
			Annotations: Annotations(),
		},
		Data: make(map[string]string),
	}
	argsLabels(configmap.Labels, d.Args[1:])
	entry(d, configmap.Entry, func(body *config.Directive) {
		asserts.ArgsLen(body, 1)
		configmap.Data[body.Name] = body.Args[0]
	})

	kube.Objects = append(kube.Objects, configmap)
}

func init() {
	kinds["configmap"] = configmapParse
	kinds["ConfigMap"] = configmapParse
}
