package refs

import (
	"github.com/ihaiker/vik8s/reduce/config"
	"reflect"
)

type (
	Handler func(fieldType reflect.Type, item *config.Directive) interface{}
	Manager map[reflect.Type]Handler
)

func (h *Manager) DealWith(fieldType reflect.Type, item *config.Directive) (reflect.Value, bool) {
	if handler, has := (*h)[fieldType]; has {
		v := handler(fieldType, item)
		return reflect.ValueOf(v), true
	}
	return reflect.Value{}, false
}

func (h *Manager) With(fieldType reflect.Type, handler Handler) *Manager {
	(*h)[fieldType] = handler
	if fieldType.Kind() == reflect.Ptr {
		(*h)[fieldType.Elem()] = handler
	} else {
		(*h)[reflect.PtrTo(fieldType)] = handler
	}
	return h
}
