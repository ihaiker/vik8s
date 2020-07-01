package volumes

import (
	"github.com/ihaiker/vik8s/reduce/config"
	"github.com/ihaiker/vik8s/reduce/refs"
	v1 "k8s.io/api/core/v1"
)

func othersParse(volume *v1.Volume,
	volumeType, source string, args []string, body config.Directives) {

	switch volumeType {
	case "gluster":
		volumeType = "glusterfs"
	case "ceph":
		volumeType = "cephfs"
	}

	refs.Unmarshal(&volume.VolumeSource, &config.Directive{
		Name: volumeType, Args: args, Body: body,
	})
}
