package kube

import (
	"github.com/ihaiker/vik8s/reduce/asserts"
	"github.com/ihaiker/vik8s/reduce/config"
)

type Secret struct {
	*Entry
	Type string
	Data Data
}

func (s *Secret) ToYaml(indent int) string {
	w := config.Writer(indent)
	w.Writer(s.Entry.Yaml(indent))
	w.Line("type:", s.Type)
	w.Writer(s.Data.ToYaml(indent))
	return w.String()
}

func secret(d *config.Directive, kube *Kubernetes) {
	asserts.ArgsRange(d, 1, 2)

	secret := &Secret{
		Entry: &Entry{
			Name:   d.Args[0],
			Labels: Labels(), Annotations: Annotations(),
		},
		Type: d.Args[1],
		Data: make(map[string]string),
	}
	entry(d, secret.Entry, func(body *config.Directive) {
		asserts.ArgsLen(body, 1)
		secret.Data[body.Name] = body.Args[0]
	})
	kube.Objects = append(kube.Objects, secret)
}

func init() {
	kinds["secret"] = secret
	kinds["Secret"] = secret
}
