package container

import (
	"fmt"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/reduce/config"
	"github.com/ihaiker/vik8s/reduce/kube/pod/volumes"
	v1 "k8s.io/api/core/v1"
)

func mountParse(args []string, body *config.Directive, spec *v1.PodSpec, c *v1.Container) {
	vt, name, _ := volumes.VolumeTypeAndNameAndSource(args[0], args[1:])

	vm := v1.VolumeMount{Name: name}
	switch vt {
	case "from":
		vm.MountPath, vm.SubPath = utils.Split2(args[1], ":")
		c.VolumeMounts = append(c.VolumeMounts, vm)
		return
	case "empty", "emptyDir", "emptydir":
		vm.MountPath = args[1]
	case "hostpath", "hostPath":
		sourcePath := ""
		sourcePath, vm.MountPath = utils.Split2(args[1], ":")
		if vm.MountPath == "" {
			vm.MountPath = sourcePath
		}
		args[1] = sourcePath
	case "secret", "configmap", "config", "configMap":
		vm.MountPath, vm.SubPath = utils.Split2(args[1], ":")
		args = append(args[0:1], args[2:]...)
	default:
		vm.MountPath, vm.SubPath = utils.Split2(args[1], ":")
		args[0] = fmt.Sprintf("%s:%s", vt, name)
	}

	if d := body.Body.Remove("mountPropagation"); d != nil {
		mp := v1.MountPropagationMode(d.Args[0])
		vm.MountPropagation = &mp
	}
	if d := body.Body.Remove("subPath"); d != nil {
		vm.SubPath = d.Args[0]
	}
	if d := body.Body.Remove("subPathExpr"); d != nil {
		vm.SubPathExpr = d.Args[0]
	}
	if d := body.Body.Remove("readOnly"); d != nil {
		vm.ReadOnly = d.Args[0] == "true"
	}

	c.VolumeMounts = append(c.VolumeMounts, vm)
	volumes.VolumeParse(&config.Directive{Name: "volume", Args: args, Body: body.Body}, spec)
}
