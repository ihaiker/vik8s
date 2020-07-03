package kube

import (
	"github.com/ihaiker/vik8s/reduce/asserts"
	"github.com/ihaiker/vik8s/reduce/config"
	"github.com/ihaiker/vik8s/reduce/plugins"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

func out(meta metav1.TypeMeta, om metav1.ObjectMeta) string {
	w := config.Writer(0)
	w.Line("apiVersion:", meta.APIVersion)
	w.Line("kind:", meta.Kind)
	w.Line("metadata:")
	w.Indent(1).Line("name:", om.Name)
	if om.Namespace != "" {
		w.Indent(1).Line("namespace:", om.Namespace)
	}
	if len(om.Labels) > 0 {
		w.Indent(1).Line("labels:")
		for label, value := range om.Labels {
			w.Indent(2).Writer(label, ": ")
			w.Writer(value)
			w.Enter()
		}
	}
	if len(om.Annotations) > 0 {
		w.Indent(1).Line("annotations:")
		for label, value := range om.Annotations {
			w.Indent(2).Writer(label, ": ", value).Enter()
		}
	}
	return w.String()
}

func configMapToString(configMap *v1.ConfigMap) string {
	w := config.Writer(0)
	w.Writer(out(configMap.TypeMeta, configMap.ObjectMeta))

	if len(configMap.Data) > 0 {
		w.Line("data:")
		for label, value := range configMap.Data {
			if strings.Index(value, "\n") == -1 {
				w.Indent(1).Writer(label, ": ", value).Enter()
			} else {
				w.Indent(1).Writer(label, ": |-").Enter()
				w.Writer(config.ToString([]byte(value), 2))
			}
		}
	}
	return w.String()
}

func configMapParse(version, prefix string, directive *config.Directive) metav1.Object {
	asserts.ArgsMin(directive, 1)
	configMap := &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: v1.SchemeGroupVersion.String(),
		},
	}
	asserts.Metadata(configMap.GetObjectMeta(), directive)
	configMap.Data = make(map[string]string)
	for _, d := range directive.Body {
		configMap.Data[d.Name] = d.Args[0]
	}
	return configMap
}

var ConfigMap = plugins.ReduceHandler{
	Names: []string{"configmap", "config", "ConfigMap"}, Handler: configMapParse,
	Demo: `
configmap data-config [label1=value1 ...] {
    datakey datavalue;
    nginx.conf '
    http {
        server {

        }
    }
    ';
    password haiker:abd123123123;
}

configmap data-config-2 {
	labels {
		label1 value1;
	}
	data-key-1 value-1;
	data-key-2 value-2;
}
`,
}
