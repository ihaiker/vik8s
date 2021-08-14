package tools

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"github.com/ihaiker/vik8s/install/paths"
	"github.com/ihaiker/vik8s/libs/utils"
	yamls "github.com/ihaiker/vik8s/yaml"
	"io"
	"os"
	"regexp"
	"strings"
	"text/template"
)

var systemFuncs = template.FuncMap{
	"strjoin": strings.Join,
	"base64": func(str ...string) string {
		outs := bytes.NewBufferString("")
		for _, s := range str {
			outs.WriteString(s)
		}
		return base64.StdEncoding.EncodeToString(outs.Bytes())
	},
	"environ": func(env, def string) string {
		if v := os.Getenv(env); v != "" {
			return v
		}
		return def
	},
	"indent": func(n int, text string) string {
		startOfLine := regexp.MustCompile(`(?m)^`)
		indentation := strings.Repeat(" ", n)
		return startOfLine.ReplaceAllLiteralString(text, indentation)
	},
}

func Template(content string, data interface{}, funcs ...template.FuncMap) ([]byte, error) {
	tempFuns := template.FuncMap{}
	for k, f := range systemFuncs {
		tempFuns[k] = f
	}
	for _, funcMap := range funcs {
		for k, f := range funcMap {
			tempFuns[k] = f
		}
	}

	out := bytes.NewBufferString("")
	t, err := template.New("").Funcs(tempFuns).Parse(content)
	if err != nil {
		return nil, utils.Wrap(err, "template error")
	}

	if err = t.Execute(out, data); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

//Assert Search assert file. First, it will search user local home path $HOME/.vik8s/$cloud/$name.
//if not found, then the default name file will return
func Assert(name string, data interface{}, funcs ...template.FuncMap) ([]byte, error) {
	var tmpCtx []byte
	var err error

	localFile := paths.Join(name)
	if utils.Exists(localFile) {
		tmpCtx = utils.FileBytes(localFile)
	} else if tmpCtx, err = yamls.Asset(name); err != nil {
		return nil, err
	}

	var outs []byte
	if outs, err = Template(string(tmpCtx), data, funcs...); err != nil {
		return nil, err
	}

	//去除空行，使得的文件好看一下吧
	pretty := bytes.NewBuffer([]byte{})
	reader := bufio.NewReader(bytes.NewReader(outs))

	var line []byte
	var isPrefix bool
	for {
		line, isPrefix, err = reader.ReadLine()
		if err == io.EOF {
			err = nil
			break
		}
		if !isPrefix && strings.TrimSpace(string(line)) == "" {
			continue
		}
		pretty.Write(line)
		if !isPrefix {
			pretty.WriteRune('\n')
		}
	}
	return pretty.Bytes(), err
}
