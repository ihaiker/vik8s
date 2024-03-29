package kube

import (
	"bufio"
	"bytes"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/reduce/config"
	"github.com/ihaiker/vik8s/reduce/plugins"
	"io"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"path/filepath"
	"reflect"
	"sigs.k8s.io/yaml"
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

func removeStatus(bs []byte) string {
	bs = bytes.ReplaceAll(bs, []byte("status: {}\n"), []byte{})

	readLine := func(r *bufio.Reader) (string, error) {
		outs := bytes.NewBufferString("")
		for {
			if line, prefix, err := r.ReadLine(); err != nil {
				return "", err
			} else if prefix {
				outs.Write(line)
			} else {
				outs.Write(line)
				return outs.String(), err
			}
		}
	}

	outs := bytes.NewBufferString("")
	reader := bufio.NewReader(bytes.NewBuffer(bs))
	statusStart := false
	for {
		if line, err := readLine(reader); err == io.EOF {
			break
		} else {
			if statusStart {
				if line[0:1] != " " {
					statusStart = false
				} else {
					continue
				}
			}
			if line == "status:" {
				statusStart = true
				continue
			}
			outs.WriteString(line)
			outs.WriteRune('\n')
		}
	}
	return outs.String()
}

func fix(bs []byte) string {
	bs = bytes.ReplaceAll(bs, []byte("  creationTimestamp: null\n"), []byte{})
	bs = bytes.ReplaceAll(bs, []byte("spec: {}\n"), []byte{})
	//fix 这个标签会有些问题
	bs = bytes.ReplaceAll(bs, []byte("          labels:\n"), []byte("      labels:\n"))
	return removeStatus(bs)
}

func (k *Kubernetes) String() string {
	return string(k.Bytes())
}

func (k *Kubernetes) Bytes() []byte {
	w := config.Writer(0).
		Line("# -------------------------------------- #").
		Line("#          Generate by vik8s             #").
		Format("#       Kubernetes version %-8s      #\n", k.Kubernetes).
		Line("#    https://github.com/ihaiker/vik8s    #").
		Line("# -------------------------------------- #")

	for _, object := range k.Objects {
		w.Line("---")

		if ns, match := object.(*v1.Namespace); match {
			ns.SetNamespace("")
		} else if ns := object.GetNamespace(); ns == "" {
			object.SetNamespace(k.Namespace())
		}

		switch t := object.(type) {
		case *sourceYaml:
			w.Writer(t.Data).Enter()
		case *v1.ConfigMap:
			w.Writer(configMapToString(t))
		case *v1.Secret:
			w.Writer(secretToString(t))
		default:
			bs, err := yaml.Marshal(object)
			utils.Panic(err, "Marshal error %s", reflect.TypeOf(object).String())
			w.Writer(fix(bs)).Enter()
		}
		w.Enter()
	}
	return w.Bytes()
}

func ParseWith(bs []byte) *Kubernetes {
	cfg := config.MustParseWith("", bs)
	return parse(cfg)
}

func Reduce(filename string) *Kubernetes {
	filePath, _ := filepath.Abs(filename)
	cfg := config.MustParse(filePath)
	return parse(cfg)
}

func parse(cfg *config.Directive) *Kubernetes {
	kube := &Kubernetes{
		Kubernetes: "v1.18.2", Prefix: "apps.vik8s.io",
		Objects: make([]metav1.Object, 0),
	}
	plugins.Load()

	if d := cfg.Body.Remove("kubernetes"); d != nil {
		kube.Kubernetes = d.Args[0]
	}
	if d := cfg.Body.Remove("prefix"); d != nil {
		kube.Prefix = d.Args[0]
	}

	replace(cfg)

	for _, d := range cfg.Body {
		if obj, handler := plugins.Manager.Handler(kube.Kubernetes, kube.Prefix, d); handler {
			kube.Objects = append(kube.Objects, obj)
		} else if obj, handler := ReduceKinds.Handler(kube.Kubernetes, kube.Prefix, d); handler {
			kube.Objects = append(kube.Objects, obj)
		} else if object, has := kubeKinds(kube.Prefix, d); has {
			kube.Objects = append(kube.Objects, object)
		} else {
			utils.Assert(false, "not support [%s]", d.Name)
		}
	}
	return kube
}
