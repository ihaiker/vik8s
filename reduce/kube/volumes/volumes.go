package volumes

import (
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/reduce/config"
)

type (
	Volume interface {
		Name() string
		ToYaml(indent int) string
	}
	Volumes []Volume

	VolumeMount struct {
		name             string
		mountPath        string
		mountPropagation string
		subPath          string
		subPathExpr      string
		readOnly         string
	}
	VolumeMounts []*VolumeMount

	VolumeFun func(name string, args []string, bodys config.Directives) Volume
)

func (v *VolumeMounts) ToYaml(indent int) string {
	w := config.Writer(indent)
	if len(*v) > 0 {
		w.Line("volumeMounts:")
		for _, mount := range *v {
			w.Indent(1).Line("-", "name:", mount.name)
			w.Indent(2).Line("mountPath:", mount.mountPath)

			if mount.mountPropagation != "" {
				w.Indent(2).Line("mountPropagation:", mount.mountPropagation)
			}
			if mount.subPath != "" {
				w.Indent(2).Line("subPath:", mount.subPath)
			}
			if mount.subPathExpr != "" {
				w.Indent(2).Line("subPathExpr:", mount.subPathExpr)
			}
			if mount.readOnly != "" {
				w.Indent(2).Line("readOnly:", mount.readOnly)
			}
		}
	}
	return w.String()
}

var parses = map[string]VolumeFun{}

func (vs *Volumes) Add(v Volume) {
	for _, volume := range *vs {
		if volume.Name() == v.Name() {
			switch t := v.(type) {
			case *ConfigMap:
				for k, v := range t.Items {
					volume.(*ConfigMap).Items[k] = v
				}
			case *Secret:
				for k, v := range t.Items {
					volume.(*ConfigMap).Items[k] = v
				}
			}
			return
		}
	}
	*vs = append(*vs, v)
}

func (vs *Volumes) ToYaml(indent int) string {
	w := config.Writer(indent)
	if len(*vs) > 0 {
		w.Line("volumes:")
		for _, volume := range *vs {
			w.Writer(volume.ToYaml(indent + 1))
		}
	}
	return w.String()
}

func volumeTypeAndName(args []string) (string, string) {
	vt, name := utils.Split2(args[0], ":")
	if name == "" {
		name = vt
		vt = utils.Switch(len(args) == 1, "emptyDir", "hostPath")
	}
	return vt, name
}

func VolumeParse(args []string, body []*config.Directive) Volume {
	vt, name := volumeTypeAndName(args)
	fn, has := parses[vt]
	utils.Assert(has, "not support volume %s", vt)
	return fn(name, args[1:], body)
}
