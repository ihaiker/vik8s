package volumes

import (
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/reduce/config"
	v1 "k8s.io/api/core/v1"
)

func VolumeTypeAndNameAndSource(name string, args []string) (string, string, string) {
	vt, volumeName, sourceName := utils.Split3(name, ":")
	if volumeName == "" {
		volumeName = vt
		vt = utils.Switch(len(args) == 1, "emptyDir", "hostPath")
	}
	if sourceName == "" {
		sourceName = volumeName
	}
	return vt, volumeName, sourceName
}

func volumeParse(name string, args []string, body config.Directives) v1.Volume {
	vt, volumeName, sourceName := VolumeTypeAndNameAndSource(name, args)
	volume := v1.Volume{Name: volumeName}
	switch vt {
	case "emptyDir", "emptydir", "empty":
		emptyDirParse(&volume, sourceName, args, body)
	case "hostpath", "hostPath":
		hostPathParse(&volume, sourceName, args, body)
	case "config", "configmap", "configMap":
		configMapParse(&volume, sourceName, args, body)
	case "secret":
		secretParse(&volume, sourceName, args, body)
	case "pvc":
		pvcParse(&volume, sourceName, args, body)
	default:
		othersParse(&volume, vt, sourceName, []string{}, body)
	}
	return volume
}

func VolumesParse(d *config.Directive, spec *v1.PodSpec) {
	for _, body := range d.Body {
		volume := volumeParse(body.Name, body.Args, body.Body)
		addVolume(spec, volume)
	}
}

func addVolume(spec *v1.PodSpec, volume v1.Volume) {
	for _, v := range spec.Volumes {
		if v.Name == volume.Name {
			return
		}
	}
	spec.Volumes = append(spec.Volumes, volume)
}

func VolumeParse(d *config.Directive, spec *v1.PodSpec) {
	volume := volumeParse(d.Args[0], d.Args[1:], d.Body)
	addVolume(spec, volume)
}
