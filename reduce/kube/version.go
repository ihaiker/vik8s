package kube

import (
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/reduce/refs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"path/filepath"
	"reflect"
	"strings"
)

type Version struct {
	Kubernetes string
}

func (v Version) Set(obj metav1.Object) {
	meta := metav1.TypeMeta{
		Kind: filepath.Ext(reflect.TypeOf(obj).String())[1:],
	}
	meta.APIVersion = reflect.TypeOf(obj).Elem().PkgPath()
	if strings.HasPrefix(meta.APIVersion, "k8s.io/api/core/") {
		meta.APIVersion = meta.APIVersion[16:]
	} else if strings.HasPrefix(meta.APIVersion, "k8s.io/api/") {
		meta.APIVersion = meta.APIVersion[11:]
	}
	err := refs.SetField(obj, "TypeMeta", meta)
	utils.Panic(err, "Set TypeMeta")
}
