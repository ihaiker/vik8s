package utils

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
)

func Safe(fn func()) (err error) {
	Try(fn, func(e error) {
		err = e
	})
	return
}

//Try handler(err)
func Try(fun func(), handler ...func(error)) {
	defer Catch(handler...)
	fun()
}

//Try handler(err) and finally
func TryFinally(fun func(), handler func(error), finallyFn func()) {
	defer func() {
		if finallyFn != nil {
			finallyFn()
		}
	}()
	Try(fun, handler)
}

type WrapError struct {
	Err     error
	Message string
}

func (w WrapError) Error() string {
	return fmt.Sprintf("%s: %s", w.Message, w.Err)
}

func Wrap(err error, format string, param ...interface{}) error {
	if _, match := err.(*WrapError); match {
		return err
	} else {
		return &WrapError{
			Err: err, Message: fmt.Sprintf(format, param...),
		}
	}
}

func Catch(fns ...func(error)) {
	if r := recover(); r != nil && len(fns) > 0 {
		if err, match := r.(error); match {
			for _, fn := range fns {
				fn(err)
			}
		} else {
			err := fmt.Errorf("%v", r)
			for _, fn := range fns {
				fn(err)
			}
		}
	}
}

func Assert(check bool, msg string, params ...interface{}) {
	if !check {
		panic(&WrapError{Err: fmt.Errorf(msg, params...), Message: "Assert"})
	}
}

//如果不为空，使用msg panic错误，
func Panic(err error, msg string, params ...interface{}) {
	if err != nil {
		panic(Wrap(err, fmt.Sprintf(msg, params...)))
	}
}

func Stack() string {
	stackBuf := make([]uintptr, 50)
	length := runtime.Callers(3, stackBuf[:])
	stack := stackBuf[:length]
	trace := ""
	frames := runtime.CallersFrames(stack)
	for {
		frame, more := frames.Next()
		if strings.HasSuffix(frame.File, "/vik8s/libs/utils/errors.go") ||
			strings.HasSuffix(frame.File, "/src/runtime/panic.go") ||
			strings.HasSuffix(frame.File, "/testing/testing.go") ||
			frame.Function == "runtime.goexit" || frame.Function == "" {

		} else if strings.HasPrefix(frame.Function, "github.com/ihaiker/vik8s") {
			trace = trace + fmt.Sprintf("\t%s:%d %s\n", frame.File, frame.Line, filepath.Base(frame.Function))
		}

		if !more {
			break
		}
	}
	return trace
}
