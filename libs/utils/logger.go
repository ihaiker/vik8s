package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

func Logs(format string, args ...interface{}) {
	fmt.Printf(format, args...)
	fmt.Println()
}

func Line(format string, params ...interface{}) {
	out := bytes.NewBuffer([]byte{})
	_, _ = fmt.Fprint(out, strings.Repeat("- ", 15))
	_, _ = fmt.Fprintf(out, format, params...)
	_, _ = fmt.Fprint(out, strings.Repeat(" -", 15))
	out.WriteByte('\n')
	_, _ = os.Stdout.Write(out.Bytes())
}

func Stdout(name string) func(reader io.Reader) error {
	return func(reader io.Reader) error {
		r := bufio.NewReader(reader)
		for {
			line, isPrefix, err := r.ReadLine()
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return err
			}
			if isPrefix {
				fmt.Print(string(line))
			} else {
				fmt.Println(name, string(line))
			}
		}
	}
}
