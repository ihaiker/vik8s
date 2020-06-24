package tools

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"github.com/ihaiker/vik8s/libs/ssh"
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

func Template(content string, data interface{}, funcs ...template.FuncMap) *bytes.Buffer {
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
	utils.Panic(err, "template error")
	err = t.Execute(out, data)
	utils.Panic(err, "template error")
	return out
}

func MustAssert(name string, data interface{}, funcs ...template.FuncMap) []byte {
	localFile := Join(name)

	var templateFile []byte
	if utils.Exists(localFile) {
		templateFile = utils.FileBytes(localFile)
	} else {
		templateFile = yamls.MustAsset(name)
	}

	out := Template(string(templateFile), data, funcs...)

	//去除空行，使得的文件好看一下吧
	pretty := bytes.NewBuffer([]byte{})
	reader := bufio.NewReader(out)
	for {
		line, isPrefix, err := reader.ReadLine()
		if err == io.EOF {
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
	return pretty.Bytes()
}

func MustScpAndApplyAssert(node *ssh.Node, name string, data interface{}, funcs ...template.FuncMap) {
	pods := MustAssert(name, data, funcs...)
	remote := node.Vik8s("apply", strings.TrimPrefix(name, "yaml/"))

	err := node.ScpContent(pods, remote)
	utils.Panic(err, "scp %s", name)

	err = node.CmdStd("kubectl apply -f "+remote, os.Stdout)
	utils.Panic(err, "kubectl apply -f %s", remote)
}
