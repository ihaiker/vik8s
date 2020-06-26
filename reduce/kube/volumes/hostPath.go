package volumes

import (
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/reduce/config"
)

type HostPath struct {
	name     string
	path     string
	fileType string
}

func (e *HostPath) ToYaml(indent int) string {
	w := config.Writer(indent)
	w.Line("- name:", e.name)
	w.Indent(1).Line("hostPath:")
	w.Indent(2).Line("path:", e.path)
	if e.fileType != "" {
		w.Indent(2).Line("type:", e.fileType)
	}
	return w.String()
}

func (e *HostPath) Name() string {
	return e.name
}

func hostPathParse(name string, args []string, bodys config.Directives) Volume {
	v := &HostPath{name: name}
	v.path = args[0]
	v.fileType = utils.Index(args, 1)
	return v
}

func init() {
	parses["hostPath"] = hostPathParse
	parses["hostpath"] = hostPathParse
}
