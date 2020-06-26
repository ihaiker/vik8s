package kube

import (
	"github.com/ihaiker/vik8s/reduce/asserts"
	"github.com/ihaiker/vik8s/reduce/config"
	"path/filepath"
	"reflect"
)

var kinds = map[string]func(d *config.Directive, kube *Kubernetes){}

type (
	Kubernetes struct {
		Kubernetes string
		Prefix     string
		Objects    []YamlEntry
	}
)

func (k *Kubernetes) Add(object YamlEntry) {
	k.Objects = append(k.Objects, object)
}

func (k *Kubernetes) String() string {
	w := config.Writer(0).Line("# generate by vik8s")

	namespace := ""
	for _, obj := range k.Objects {
		switch obj.(type) {
		case *Namespace:
			namespace = obj.(*Namespace).Name
		}
	}
	version := Version{k.Kubernetes}
	for _, object := range k.Objects {
		w.Line("---")
		kind := filepath.Ext(reflect.TypeOf(object).String())[1:]
		object.(IEntry).Set(namespace, kind, version.Get(kind))
		w.Writer(object.ToYaml(0))
		w.Enter()
	}
	return w.String()
}

func init() {
	kinds["kubernetes"] = func(d *config.Directive, kube *Kubernetes) {
		asserts.ArgsLen(d, 1)
		kube.Kubernetes = d.Args[0]
	}
	kinds["prefix"] = func(d *config.Directive, kube *Kubernetes) {
		asserts.ArgsLen(d, 1)
		kube.Prefix = d.Args[0]
	}
}

func Parse(filename string) *Kubernetes {
	kube := &Kubernetes{
		Kubernetes: "1.18.2", Prefix: "vik8s.io",
		Objects: make([]YamlEntry, 0),
	}
	f, _ := filepath.Abs(filename)
	cfg := config.MustParse(f)
	for _, d := range cfg.Body {
		for kindName, kind := range kinds {
			if kindName == d.Name {
				kind(d, kube)
			}
		}
	}
	return kube
}
