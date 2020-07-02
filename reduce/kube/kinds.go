package kube

import (
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/reduce/asserts"
	"github.com/ihaiker/vik8s/reduce/config"
	"github.com/ihaiker/vik8s/reduce/refs"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/api/rbac/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
)

var schemes = runtime.NewScheme()

func init() {
	_ = v1beta1.AddToScheme(schemes)
	_ = v1.AddToScheme(schemes)
}

func kubeKinds(prefix string, item *config.Directive) (metav1.Object, bool) {
	kind, version := utils.Split2(item.Name, ":")
	for knownKind, knownType := range schemes.AllKnownTypes() {
		if knownKind.String() == version || knownKind.Kind == kind {
			objValue := reflect.New(knownType)
			obj := objValue.Interface().(metav1.Object)

			typeMeta := objValue.Elem().FieldByName("TypeMeta")
			typeMeta.Set(reflect.ValueOf(metav1.TypeMeta{
				Kind: kind, APIVersion: version,
			}))

			asserts.Metadata(obj, item)
			asserts.AutoLabels(obj, prefix)
			for _, directive := range item.Body {
				refs.Unmarshal(obj, directive)
			}
			return obj, true
		}
	}
	return nil, false
}
