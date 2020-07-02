package service

import (
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/reduce/asserts"
	"github.com/ihaiker/vik8s/reduce/config"
	"github.com/ihaiker/vik8s/reduce/refs"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"math"
)

func ServiceParse(version, prefix string, dir *config.Directive) []metav1.Object {
	service := &v1.Service{}
	asserts.MetadataIndex(service, dir, math.MaxInt8)
	asserts.AutoLabels(service, prefix)

	if serviceType := utils.Index(dir.Args, 1); serviceType != "" {
		service.Spec.Type = v1.ServiceType(serviceType)
	}

	for _, item := range dir.Body {
		switch item.Name {
		default:
			refs.Unmarshal(&service.Spec, item)

		case "port":
			service.Spec.Ports = append(service.Spec.Ports, servicePortParse(item.Args))
		case "ports":
			for _, i := range item.Body {
				service.Spec.Ports = append(service.Spec.Ports,
					servicePortParse(append([]string{i.Name}, i.Args...)))
			}
		}
	}
	return []metav1.Object{service}
}

func servicePortParse(args []string) v1.ServicePort {
	targetPort, portAndProtocol := utils.Split2(args[1], ":")
	port, protocol := utils.Split2(portAndProtocol, "/")
	sp := v1.ServicePort{
		Name: args[0], Protocol: v1.Protocol(protocol),
		Port:       *utils.Int32(port, 10),
		TargetPort: intstr.Parse(targetPort),
		NodePort:   *utils.Int32(utils.Index(args, 2), 10),
	}
	return sp
}
