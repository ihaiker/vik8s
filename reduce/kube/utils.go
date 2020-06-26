package kube

import (
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/reduce/asserts"
	"github.com/ihaiker/vik8s/reduce/config"
)

func argsLabels(labelMap *Properties, labels []string) {
	for _, arg := range labels {
		label, value := utils.Split2(arg, "=")
		labelMap.Add(label, value)
	}
}

func bodyLabels(labelMap *Properties, d *config.Directive) {
	for _, body := range d.Body {
		asserts.ArgsLen(body, 1)
		labelMap.Add(body.Name, body.Args[0])
	}
}

func annotations(an *Properties, d *config.Directive) {
	for _, body := range d.Body {
		asserts.ArgsLen(body, 1)
		an.Add(body.Name, body.Args[0])
	}
}
