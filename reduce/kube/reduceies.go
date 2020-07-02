package kube

import (
	"github.com/ihaiker/vik8s/reduce/config"
	"github.com/ihaiker/vik8s/reduce/kube/daemonset"
	"github.com/ihaiker/vik8s/reduce/kube/deployment"
	"github.com/ihaiker/vik8s/reduce/kube/ingress"
	"github.com/ihaiker/vik8s/reduce/kube/pod"
	"github.com/ihaiker/vik8s/reduce/kube/service"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type (
	KindMaker func(version, prefix string, directive *config.Directive) []metav1.Object
)

type sourceYaml struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Data              string
}

func yamlParse(version, prefix string, directive *config.Directive) []metav1.Object {
	v := &sourceYaml{}
	v.Data = directive.Args[0]
	return []metav1.Object{v}
}

var reduceKinds = map[string]KindMaker{
	"namespace": namespaceParse, "node": nodeParse,
	"configmap": configMapParse, "secret": secretParse,

	"pod": pod.Parse,

	"dep": deployment.Parse, "deployment": deployment.Parse, "Deployment": deployment.Parse,
	"daemon": daemonset.Parse, "daemonset": daemonset.Parse, "DaemonSet": daemonset.Parse,

	"service": service.ServiceParse, "Service": service.ServiceParse,
	"Ingress": ingress.Parse, "ingress": ingress.Parse,

	"yaml": yamlParse,
}
