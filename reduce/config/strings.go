package config

import (
	"bytes"
	"fmt"
	"strings"
)

type writer struct {
	outs   *bytes.Buffer
	prefix string
}

func Writer(indent int) *writer {
	return &writer{
		prefix: strings.Repeat(" ", indent*2),
		outs:   bytes.NewBufferString(""),
	}
}

func (w *writer) Tab() *writer {
	return w.Indent(1)
}

func (w *writer) Indent(indent int) *writer {
	w.outs.WriteString(strings.Repeat(" ", indent*2))
	return w
}

func (w *writer) Format(format string, param ...interface{}) *writer {
	w.outs.WriteString(w.prefix)
	w.outs.WriteString(fmt.Sprintf(format, param...))
	return w
}

func (w *writer) Line(args ...string) *writer {
	w.outs.WriteString(w.prefix)
	for i, arg := range args {
		w.outs.WriteString(arg)
		if i < len(args)-1 {
			w.outs.WriteString(" ")
		}
	}
	return w.Enter()
}

func (w *writer) Writer(args ...string) *writer {
	for _, arg := range args {
		w.outs.WriteString(arg)
	}
	return w
}

func (w *writer) Enter() *writer {
	w.outs.WriteRune('\n')
	return w
}

func (s *writer) String() string {
	return s.outs.String()
}
