package refs

import (
	"github.com/ihaiker/vik8s/reduce/config"
	"reflect"
)

type (
	TypeHandler func(fieldType reflect.Type, item *config.Directive) interface{}
	TypeManager map[reflect.Type]TypeHandler
)

func (h *TypeManager) DealWith(fieldType reflect.Type, item *config.Directive) (reflect.Value, bool) {
	if handler, has := (*h)[fieldType]; has {
		v := handler(fieldType, item)
		return reflect.ValueOf(v), true
	}
	return reflect.Value{}, false
}

func (h *TypeManager) With(fieldType reflect.Type, handler TypeHandler) *TypeManager {
	(*h)[fieldType] = handler
	if fieldType.Kind() == reflect.Ptr {
		(*h)[fieldType.Elem()] = handler
	} else {
		(*h)[reflect.PtrTo(fieldType)] = handler
	}
	return h
}
