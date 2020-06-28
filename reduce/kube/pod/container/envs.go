package container

import (
	"github.com/ihaiker/vik8s/libs/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func envParse(name string, args []string) v1.EnvVar {
	env := v1.EnvVar{Name: name}

	if len(args) == 1 {
		env.Value = args[0]
	} else {
		env.ValueFrom = &v1.EnvVarSource{}

		switch args[0] {
		case "field":
			env.ValueFrom.FieldRef = &v1.ObjectFieldSelector{
				FieldPath: args[1],
			}
		case "configMap", "config", "configmap":
			env.ValueFrom.ConfigMapKeyRef = &v1.ConfigMapKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: args[1],
				},
				Key: args[2],
			}
		case "secret":
			env.ValueFrom.SecretKeyRef = &v1.SecretKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: args[1],
				},
				Key: args[2],
			}
		case "resource", "res":
			env.ValueFrom.ResourceFieldRef = &v1.ResourceFieldSelector{
				ContainerName: args[1],
				Resource:      args[2],
				Divisor:       resource.MustParse(utils.Default(args, 3, "1")),
			}
		}
	}
	return env
}
