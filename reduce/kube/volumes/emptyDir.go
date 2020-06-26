package volumes

import (
	"github.com/ihaiker/vik8s/reduce/config"
)

type EmptyDir struct {
	name      string
	medium    string
	sizeLimit string
}

func (e *EmptyDir) ToYaml(indent int) string {
	w := config.Writer(indent)
	w.Line("- name:", e.name)
	if e.medium == "" && e.sizeLimit == "" {
		//w.Line("  emptyDir:")
		//w.Indent(2).Line("{}")
	} else {
		w.Line("  emptyDir:")
		if e.medium != "" {
			w.Indent(2).Line("medium:", e.medium)
		}
		if e.sizeLimit != "" {
			w.Indent(2).Line("sizeLimit:", e.sizeLimit)
		}
	}
	return w.String()
}

func (e *EmptyDir) Name() string {
	return e.name
}

func emptyDirParse(name string, args []string, bodys config.Directives) Volume {
	e := &EmptyDir{name: name}
	if d := bodys.Get("medium"); d != nil {
		e.medium = d.Args[0]
	}
	if d := bodys.Get("sizeLimit"); d != nil {
		e.sizeLimit = d.Args[0]
	}
	return e
}

func init() {
	parses["emptyDir"] = emptyDirParse
	parses["emptydir"] = emptyDirParse
	parses["empty"] = emptyDirParse
}
