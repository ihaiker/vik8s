package volumes

import (
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/reduce/config"
)

type persistentVolumeClaim struct {
	name      string
	claimName string
	readonly  string
}

func (p *persistentVolumeClaim) Name() string {
	return p.name
}

func (p *persistentVolumeClaim) ToYaml(indent int) string {
	w := config.Writer(indent)
	w.Line("- name:", p.name)
	w.Indent(1).Line("persistentVolumeClaim:")
	w.Indent(2).Line("claimName:", p.claimName)
	if p.readonly == "true" {
		w.Indent(2).Line("readOnly: true")
	}
	return w.String()
}

func pvcParse(name string, args []string, bodys config.Directives) Volume {
	pvc := &persistentVolumeClaim{}
	pvc.name, pvc.readonly = utils.Split2(name, ":")
	if pvc.claimName = utils.Index(args, 0); pvc.claimName == "" {
		pvc.claimName = pvc.name
	}
	return pvc
}

func init() {
	parses["pvc"] = pvcParse
}
