package utils

import (
	"reflect"
	"sync"
)

type async struct {
	runs []func()
	gw   *sync.WaitGroup
}

func Async() *async {
	return &async{
		runs: make([]func(), 0),
		gw:   new(sync.WaitGroup),
	}
}

func (as *async) Add(fn interface{}, params ...interface{}) {
	as.gw.Add(1)
	go func() {
		defer as.gw.Done()

		in := make([]reflect.Value, 0)
		for _, param := range params {
			in = append(in, reflect.ValueOf(param))
		}

		reflect.ValueOf(fn).Call(in)
	}()
}

func (as *async) Wait() {
	as.gw.Wait()
}
