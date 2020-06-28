package config

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math"
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

func ToString(bs []byte, indent int) string {
	reader := bufio.NewReader(bytes.NewBuffer(bs))

	readLine := func(r *bufio.Reader) (string, error) {
		outs := bytes.NewBufferString("")
		for {
			if line, prefix, err := r.ReadLine(); err != nil {
				return "", err
			} else if prefix {
				outs.Write(line)
			} else {
				outs.Write(line)
				return outs.String(), err
			}
		}
	}

	min := math.MaxInt8
	for {
		if line, err := readLine(reader); err == nil {
			newline := strings.TrimLeft(line, " ")
			if newline == "" {
				continue
			}
			space := len(line) - len(newline)
			if space < min {
				min = space
			}
		} else if err == io.EOF {
			break
		}
	}

	first := true
	w := Writer(indent)
	reader = bufio.NewReader(bytes.NewBuffer(bs))
	for {
		if line, err := readLine(reader); err == nil {
			if strings.Trim(line, " ") == "" {
				if !first {
					w.Enter()
				}
			} else {
				w.Indent(indent).Writer(line[min:]).Enter()
			}
			first = false
		} else if err == io.EOF {
			break
		}
	}

	return w.String()
}
