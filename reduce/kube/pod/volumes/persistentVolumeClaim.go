package volumes

import (
	"github.com/ihaiker/vik8s/reduce/config"
	v1 "k8s.io/api/core/v1"
)

func pvcParse(v *v1.Volume,
	source string, args []string, bodys config.Directives) {
	v.PersistentVolumeClaim = &v1.PersistentVolumeClaimVolumeSource{
		ClaimName: source, ReadOnly: false,
	}
	if d := bodys.Remove("readOnly"); d != nil {
		v.PersistentVolumeClaim.ReadOnly = d.Args[0] == "true"
	}
}
