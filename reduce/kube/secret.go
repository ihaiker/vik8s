package kube

import (
	"bytes"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/reduce/asserts"
	"github.com/ihaiker/vik8s/reduce/config"
	"github.com/ihaiker/vik8s/reduce/plugins"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func secretToString(secret *v1.Secret) string {
	w := config.Writer(0)
	w.Writer(out(secret.TypeMeta, secret.ObjectMeta))

	w.Line("data:")
	for label, value := range secret.Data {
		if bytes.IndexByte(value, '\n') == -1 {
			w.Indent(1).Writer(label, ": ", string(value)).Enter()
		} else {
			w.Indent(1).Writer(label, ": |-").Enter()
			w.Writer(config.ToString(value, 4))
		}
	}
	return w.String()
}

func secretParse(version, prefix string, directive *config.Directive) metav1.Object {
	asserts.ArgsMin(directive, 1)

	secret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: v1.SchemeGroupVersion.String(),
		},
	}
	asserts.MetadataIndex(secret.GetObjectMeta(), directive, 2)

	if st := utils.Index(directive.Args, 1); st != "" {
		secret.Type = v1.SecretType(st)
	}
	secret.Data = make(map[string][]byte)
	for _, d := range directive.Body {
		secret.Data[d.Name] = []byte(d.Args[0])
	}

	return secret
}

var Secret = plugins.ReduceHandler{
	Names: []string{"secret", "Secret"}, Handler: secretParse,
	Demo: `
secret data-config-1 [secretType] {
    datakey ZGF0YXZhbHVlCg==;
    password aGFpa2VyOmFiZDEyMzEyMzEyMwo=;
}
secret data-config-2 {
	labels {
		label1 value1;
	}
	label label2 value2;

	data-key-1 dmFsdWUtMQo=;
	data-key-2 dmFsdWUtMgo=;
}
`,
}
