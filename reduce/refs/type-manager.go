package refs

import (
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/reduce/asserts"
	"github.com/ihaiker/vik8s/reduce/config"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	"reflect"
)

func intOrString(fieldType reflect.Type, item *config.Directive) interface{} {
	s := intstr.Parse(item.Args[0])
	if fieldType.Kind() == reflect.Ptr {
		return &s
	} else {
		return s
	}
}

func resourceListParse(fieldType reflect.Type, item *config.Directive) interface{} {
	res := v1.ResourceList{}
	for _, arg := range item.Args {
		k, v := utils.CompileSplit2(arg, ":|=")
		res[v1.ResourceName(k)] = resource.MustParse(v)
	}
	for _, directive := range item.Body {
		asserts.ArgsLen(directive, 1)
		res[v1.ResourceName(directive.Name)] = resource.MustParse(directive.Args[0])
	}
	if fieldType.Kind() == reflect.Ptr {
		return &res
	} else {
		return res
	}
}

var Defaults = TypeManager{
	reflect.TypeOf(intstr.IntOrString{}): intOrString, reflect.TypeOf(&intstr.IntOrString{}): intOrString,
	//reflect.TypeOf(v1.ResourceRequirements{}): resourceRequirementsParse, reflect.TypeOf(&v1.ResourceRequirements{}): resourceRequirementsParse,
	reflect.TypeOf(v1.ResourceList{}): resourceListParse, reflect.TypeOf(&v1.ResourceList{}): resourceListParse,
}
