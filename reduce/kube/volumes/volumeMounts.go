package volumes

import (
	"fmt"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/reduce/config"
)

func volumeType_mountName_volumeName(args []string) (string, string, string) {
	vt, mountName, volumeName := utils.Split3(args[0], ":")
	if mountName == "" {
		mountName = vt
		vt = utils.Switch(len(args) == 1, "emptyDir", "hostPath")
	}
	if volumeName == "" {
		volumeName = mountName
	}
	return vt, mountName, volumeName
}

func MountParse(args []string, body config.Directives) (*VolumeMount, Volume) {
	vt, mountName := volumeTypeAndName(args)
	switch vt {
	case "from":
		vm := &VolumeMount{name: mountName}
		set := func(name string) string {
			if d := body.Get(name); d != nil {
				return d.Args[0]
			}
			return ""
		}
		vm.mountPath, vm.subPath = utils.Split2(args[1], ":")
		vm.mountPropagation = set("mountPropagation")
		if vm.subPath == "" {
			vm.subPath = set("subPath")
		}
		vm.subPathExpr = set("subPathExpr")
		vm.readOnly = set("readOnly")
		return vm, nil
	case "empty", "emptyDir":
		vm := &VolumeMount{name: mountName}
		vm.mountPath = args[1]
		ve := VolumeParse(args, body)
		return vm, ve
	case "hostPath":
		vm := &VolumeMount{name: mountName}
		ve := VolumeParse(args, body)
		ve.(*HostPath).path, vm.mountPath = utils.Split2(args[1], ":")
		return vm, ve

	case "secret", "configMap", "configmap":
		mountPath, subPath := utils.Split2(args[1], ":")

		vm := &VolumeMount{name: mountName}
		vm.mountPath = mountPath
		vm.subPath = subPath

		args[0] = fmt.Sprintf("%s:%s", vt, mountName)
		ve := VolumeParse(args, body)
		if mountPropagation := body.Remove("mountPropagation"); mountPropagation != nil {
			vm.mountPropagation = mountPropagation.Args[0]
		}
		if vm.subPath == "" {
			if subPath := body.Remove("subPath"); subPath != nil {
				vm.subPath = subPath.Args[0]
			}
		}
		if subPathExpr := body.Remove("subPathExpr"); subPathExpr != nil {
			vm.subPathExpr = subPathExpr.Args[0]
		}
		if readOnly := body.Remove("readOnly"); readOnly != nil {
			vm.readOnly = readOnly.Args[0]
		}
		return vm, ve
	default:
		vm := &VolumeMount{name: mountName}
		vm.mountPath, vm.subPath = utils.Split2(args[1], ":")
		ve := VolumeParse(args, body)
		return vm, ve
	}
}
