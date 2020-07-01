package kube

import (
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/reduce/refs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"path/filepath"
	"reflect"
)

type Version struct {
	Kubernetes string
}

func (v Version) Set(obj metav1.Object) {
	meta := metav1.TypeMeta{
		Kind: filepath.Ext(reflect.TypeOf(obj).String())[1:],
	}
	meta.APIVersion = v.get(meta.Kind)
	err := refs.SetField(obj, "TypeMeta", meta)
	utils.Panic(err, "Set TypeMeta")
}

func (v Version) get(kind string) string {
	switch kind {
	case "Pod":
		return "v1"
	case "DaemonSet":
		return "v1"
	case "Deployment":
		return "apps/v1"
	}
	return "v1"
}
