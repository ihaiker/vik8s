package volumes

import (
	"fmt"
	"github.com/ihaiker/vik8s/reduce/config"
	"strings"
)

type Others struct {
	name       string
	volumeType string
	bodys      config.Directives
}

func (o *Others) Name() string {
	return o.name
}

func (o *Others) ToYaml(indent int) string {
	w := config.Writer(indent)

	var subOut func(appendIndent int, body *config.Directive)
	out := func(appendIndent int, body *config.Directive) {
		if len(body.Body) == 0 {
			if body.Name == "monitors" {
				w.Indent(2 + appendIndent).Line(fmt.Sprintf("%s:", body.Name))
				for _, arg := range body.Args {
					w.Indent(3+appendIndent).Line("-", arg)
				}
			} else {
				w.Indent(2+appendIndent).Line(fmt.Sprintf("%s:", body.Name),
					strings.Join(body.Args, " "))
			}
		} else {
			w.Indent(2 + appendIndent).Line(fmt.Sprintf("%s:", body.Name))
			for _, d := range body.Body {
				subOut(appendIndent+1, d)
			}
		}
	}
	subOut = out

	w.Line("- name:", o.name)
	w.Indent(1).Line(fmt.Sprintf("%s:", o.volumeType))
	for _, body := range o.bodys {
		out(0, body)
	}
	return w.String()
}

func othersParse(t string) VolumeFun {
	return func(name string, args []string, bodys config.Directives) Volume {
		return &Others{name: name, volumeType: t, bodys: bodys}
	}
}

func init() {
	for _, storage := range []string{
		"glusterfs", "cephfs", "rbd",
		"storageos", "flocker",
		"photonPersistentDisk",
		"nfs", "csi", "iscsi",
		"vsphereVolume", "awsElasticBlockStore",
		"azureDisk", "azureFile", "scaleIO",
	} {
		parses[storage] = othersParse(storage)
	}
}
