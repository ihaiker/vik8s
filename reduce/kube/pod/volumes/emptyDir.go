package volumes

import (
	"github.com/ihaiker/vik8s/reduce/config"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func emptyDirParse(volume *v1.Volume,
	source string, args []string, body config.Directives) {
	volume.EmptyDir = &v1.EmptyDirVolumeSource{}

	if d := body.Get("medium"); d != nil {
		volume.EmptyDir.Medium = v1.StorageMedium(d.Args[0])
	}
	if d := body.Get("sizeLimit"); d != nil {
		q := resource.MustParse(d.Args[0])
		volume.EmptyDir.SizeLimit = &q
	}
}
