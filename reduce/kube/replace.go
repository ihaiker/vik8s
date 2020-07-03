package kube

import (
	"bytes"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/reduce/config"
	"text/template"
)

func replace(cfg *config.Directive) {
	for itemReplace := cfg.Body.Remove("replace"); itemReplace != nil; itemReplace = cfg.Body.Remove("replace") {

		left, right := "<<", ">>"
		if delims := itemReplace.Body.Remove("@delims"); delims != nil {
			left, right = utils.Index(delims.Args, 0), utils.Index(delims.Args, 1)
			if right == "" {
				right = left
			}
		}

		args := config.Directives{}
		for {
			if r := itemReplace.Body.Remove("@args"); r != nil {
				if len(r.Args) > 0 {
					r.Name = r.Args[0]
					r.Args = r.Args[1:]
				}
				args = append(args, r)
			} else {
				break
			}
		}

		for _, arg := range args {
			out := bytes.NewBufferString("")
			templateString := itemReplace.String()
			tmpl, err := template.New("").Delims(left, right).Funcs(funcs()).Parse(templateString)
			utils.Panic(err, "template error")

			err = tmpl.Execute(out, arg)
			utils.Panic(err, "template error")

			newConfig, err := config.ParseWith("", out.Bytes())
			utils.Panic(err, "template error")

			for _, directive := range newConfig.Body {
				cfg.Body = append(cfg.Body, directive.Body...)
			}
		}
	}
}

func funcs() template.FuncMap {
	return template.FuncMap{
		"body": func(d *config.Directive, name string) *config.Directive {
			for _, directive := range d.Body {
				if directive.Name == name {
					return directive
				}
			}
			return nil
		},
		"remove": func(d *config.Directive, name string) *config.Directive {
			return d.Body.Remove(name)
		},
	}
}
