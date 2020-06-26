package kube

import (
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/reduce/asserts"
	"github.com/ihaiker/vik8s/reduce/config"
)

type (
	resources struct {
		requests struct {
			memory string
			cpu    string
		}
		limits struct {
			memory string
			cpu    string
		}
	}
)

func (r *resources) ToYaml(indent int) string {
	w := config.Writer(indent)
	w.Line("resources:")
	if r.requests.cpu != "" || r.requests.memory != "" {
		w.Indent(1).Line("requests:")
		if r.requests.cpu != "" {
			w.Indent(2).Line("cpu:", r.requests.cpu)
		}
		if r.requests.memory != "" {
			w.Indent(2).Line("memory:", r.requests.memory)
		}
	}
	if r.limits.cpu != "" || r.limits.memory != "" {
		w.Indent(1).Line("limits:")
		if r.limits.cpu != "" {
			w.Indent(2).Line("cpu:", r.limits.cpu)
		}
		if r.limits.memory != "" {
			w.Indent(2).Line("memory:", r.limits.memory)
		}
	}
	return w.String()
}

func resourceParse(d *config.Directive) *resources {
	res := &resources{}
	for _, directive := range d.Body {
		asserts.ArgsRange(directive, 1, 2)
		for _, arg := range directive.Args {
			k, v := utils.Split2(arg, "=")
			if directive.Name == "limit" {
				if k == "cpu" {
					res.limits.cpu = v
				} else {
					res.limits.memory = v
				}
			} else {
				if k == "cpu" {
					res.requests.cpu = v
				} else {
					res.requests.memory = v
				}
			}
		}
	}
	return res
}
