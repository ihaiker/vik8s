package volumes

import (
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/reduce/config"
)

type ConfigMap struct {
	name        string
	configName  string
	defaultMode string
	Items       map[string]string
}

func (c *ConfigMap) Name() string {
	return c.name
}

func (c *ConfigMap) ToYaml(indent int) string {
	w := config.Writer(indent)
	w.Line("- name:", c.name)
	w.Indent(1).Line("configMap:")
	w.Indent(2).Line("name:", c.configName)
	if c.defaultMode != "" {
		w.Indent(2).Line("defaultMode:", c.defaultMode)
	}
	if len(c.Items) > 0 {
		w.Indent(2).Line("items:")
		for k, v := range c.Items {
			w.Indent(3).Line("- key:", k)
			w.Indent(3).Line("  path:", v)
		}
	}
	return w.String()
}

func configMapParse(name string, args []string, bodys config.Directives) Volume {
	cm := &ConfigMap{}
	cm.name, cm.defaultMode = utils.Split2(name, ":")

	if cm.configName = utils.Index(args, 0); cm.configName == "" {
		cm.configName = cm.name
	}

	if len(bodys) > 0 {
		cm.Items = make(map[string]string)
		for _, body := range bodys {
			if len(body.Args) == 0 {
				cm.Items[body.Name] = body.Name
			} else {
				cm.Items[body.Name] = body.Args[0]
			}
		}
	}
	return cm
}

func init() {
	parses["configmap"] = configMapParse
	parses["configMap"] = configMapParse
	parses["config"] = configMapParse
}
