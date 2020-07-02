package refs

import (
	"github.com/ihaiker/vik8s/reduce/config"
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

var Defaults = Manager{
	reflect.TypeOf(intstr.IntOrString{}): intOrString, reflect.TypeOf(&intstr.IntOrString{}): intOrString,
}
