package volumes

import (
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/reduce/config"
	v1 "k8s.io/api/core/v1"
)

func hostPathParse(volume *v1.Volume,
	source string, args []string, body config.Directives) {
	volume.HostPath = &v1.HostPathVolumeSource{
		Path: args[0],
	}

	if hp := utils.Index(args, 1); hp != "" {
		t := v1.HostPathType(hp)
		volume.HostPath.Type = &t
	}
}
