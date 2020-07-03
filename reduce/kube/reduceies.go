package kube

import (
	"github.com/ihaiker/vik8s/reduce/config"
	"github.com/ihaiker/vik8s/reduce/kube/daemonset"
	"github.com/ihaiker/vik8s/reduce/kube/deployment"
	"github.com/ihaiker/vik8s/reduce/kube/ingress"
	"github.com/ihaiker/vik8s/reduce/kube/pod"
	"github.com/ihaiker/vik8s/reduce/kube/service"
	"github.com/ihaiker/vik8s/reduce/plugins"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type sourceYaml struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Data              string
}

func yamlParse(version, prefix string, directive *config.Directive) metav1.Object {
	v := &sourceYaml{}
	v.Data = directive.Args[0]
	return v
}

var ReduceKinds = plugins.ReduceHandlers{
	Namespace, Node, ConfigMap, Secret,
	pod.Pod, deployment.Deployment, daemonset.DaemonSet,
	service.Service, ingress.Ingress,
	{Names: []string{"yaml"}, Demo: `
yaml '
---
apiVersion: v1
kind: Pod
....
';
`, Handler: yamlParse},
}
