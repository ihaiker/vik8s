package asserts

import (
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/reduce/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Metadata(metaObject metav1.Object, directive *config.Directive) {
	MetadataIndex(metaObject, directive, 1)
}

func MetadataIndex(meta metav1.Object, directive *config.Directive, argsLabel int) {
	meta.SetName(directive.Args[0])

	{
		labels := make(map[string]string)
		if len(directive.Args) >= argsLabel {
			for _, arg := range directive.Args[argsLabel:] {
				label, value := utils.Split2(arg, "=")
				labels[label] = value
			}
		}
		//label
		for {
			if d := directive.Body.Remove("label"); d == nil {
				break
			} else {
				labels[d.Args[0]] = d.Args[1]
			}
		}
		//labels{}
		if d := directive.Body.Remove("labels"); d != nil {
			for _, ld := range d.Body {
				labels[ld.Name] = ld.Args[0]
			}
		}
		meta.SetLabels(labels)
	}

	if ns := directive.Body.Remove("namespace"); ns != nil {
		meta.SetNamespace(ns.Args[0])
	}

	{
		annotations := make(map[string]string)
		for {
			if d := directive.Body.Remove("annotation"); d == nil {
				break
			} else {
				annotations[d.Args[0]] = d.Args[1]
			}
		}
		if d := directive.Body.Remove("annotations"); d != nil {
			for _, ld := range d.Body {
				annotations[ld.Name] = ld.Args[0]
			}
		}
		meta.SetAnnotations(annotations)
	}
}
