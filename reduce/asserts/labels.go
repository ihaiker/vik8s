package asserts

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func AutoLabels(meta metav1.Object, prefix string) {
	if len(meta.GetLabels()) > 0 {
		return
	}
	meta.SetLabels(map[string]string{
		fmt.Sprintf("%s/name", prefix): meta.GetName(),
		//fmt.Sprintf("%s/kind", prefix): strings.ToLower(filepath.Ext(reflect.TypeOf(meta).String())[1:]),
	})
}
