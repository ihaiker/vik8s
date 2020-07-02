package kube

import (
	"bytes"
	"fmt"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/reduce/config"
	"github.com/ihaiker/vik8s/reduce/refs"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"path/filepath"
	"reflect"
	"sigs.k8s.io/yaml"
	"strings"
)

type (
	Kubernetes struct {
		Kubernetes string
		Prefix     string
		Objects    []metav1.Object
	}
)

func (k *Kubernetes) Add(object metav1.Object) {
	k.Objects = append(k.Objects, object)
}

func (k *Kubernetes) Namespace() string {
	for _, object := range k.Objects {
		switch ns := object.(type) {
		case *v1.Namespace:
			return ns.Name
		}
	}
	return ""
}

func (k *Kubernetes) String() string {
	w := config.Writer(0).
		Line("# -------------------------------------- #").
		Line("#          Generate by vik8s             #").
		Line(fmt.Sprintf("#       Kubernetes version %-8s      #", k.Kubernetes)).
		Line("#    https://github.com/ihaiker/vik8s    #").
		Line("# -------------------------------------- #")

	for _, object := range k.Objects {
		w.Line("---")

		k.Set(object)
		object.SetNamespace(k.Namespace())
		if ns, match := object.(*v1.Namespace); match {
			ns.SetNamespace("")
		}

		switch t := object.(type) {
		case *v1.ConfigMap:
			w.Writer(configMapToString(t))
		case *v1.Secret:
			w.Writer(secretToString(t))
		default:
			bs, err := yaml.Marshal(object)
			bs = bytes.ReplaceAll(bs, []byte("  creationTimestamp: null\n"), []byte{})
			bs = bytes.ReplaceAll(bs, []byte("status: {}\n"), []byte{})
			bs = bytes.ReplaceAll(bs, []byte("spec: {}\n"), []byte{})
			//fixbug
			bs = bytes.ReplaceAll(bs, []byte("          labels:\n"), []byte("      labels:\n"))
			utils.Panic(err, "Marshal error %s", reflect.TypeOf(object).String())
			w.Writer(string(bs)).Enter()
		}

		w.Enter()
	}
	return w.String()
}

func (k *Kubernetes) Set(obj metav1.Object) {
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

func Parse(filename string) *Kubernetes {
	kube := &Kubernetes{
		Kubernetes: "v1.18.2", Prefix: "vik8s.io",
		Objects: make([]metav1.Object, 0),
	}

	filePath, _ := filepath.Abs(filename)
	cfg := config.MustParse(filePath)

	if d := cfg.Body.Remove("kubernetes"); d != nil {
		kube.Kubernetes = d.Args[0]
	}
	if d := cfg.Body.Remove("prefix"); d != nil {
		kube.Prefix = d.Args[0]
	}

	for _, d := range cfg.Body {
		for kindName, kindHandler := range kinds {
			if kindName == strings.ToLower(d.Name) {
				objs := kindHandler(kube.Kubernetes, kube.Prefix, d)
				for _, obj := range objs {
					kube.Objects = append(kube.Objects, obj)
				}
			}
		}
	}
	return kube
}
