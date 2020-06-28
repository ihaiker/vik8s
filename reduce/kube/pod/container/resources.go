package container

import (
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/reduce/asserts"
	"github.com/ihaiker/vik8s/reduce/config"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func resourceParse(d *config.Directive) v1.ResourceRequirements {
	req := v1.ResourceRequirements{}
	for _, directive := range d.Body {
		asserts.ArgsRange(directive, 1, 2)

		res := v1.ResourceList{}
		if directive.Name == "limit" {
			req.Limits = res
		} else {
			req.Requests = res
		}

		for _, arg := range directive.Args {
			k, v := utils.Split2(arg, "=")
			res[v1.ResourceName(k)] = resource.MustParse(v)
		}
	}
	return req
}
