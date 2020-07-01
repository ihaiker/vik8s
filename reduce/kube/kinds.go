package kube

import (
	"github.com/ihaiker/vik8s/reduce/config"
	"github.com/ihaiker/vik8s/reduce/kube/deployment"
	"github.com/ihaiker/vik8s/reduce/kube/pod"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type (
	KindMaker func(version, prefix string, directive *config.Directive) metav1.Object
)

var kinds = map[string]KindMaker{
	"namespace": namespaceParse, "node": nodeParse,
	"configmap": configMapParse, "secret": secretParse,
	"pod": pod.Parse,

	"dep": deployment.Parse, "deployment": deployment.Parse, "Deployment": deployment.Parse,
}
