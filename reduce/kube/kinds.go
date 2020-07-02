package kube

import (
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/reduce/asserts"
	"github.com/ihaiker/vik8s/reduce/config"
	"github.com/ihaiker/vik8s/reduce/refs"
	admissionv1 "k8s.io/api/admission/v1"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/apps/v1"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	auditregistrationv1alpha1 "k8s.io/api/auditregistration/v1alpha1"
	authenticationv1 "k8s.io/api/authentication/v1"
	authenticationv1beta1 "k8s.io/api/authentication/v1beta1"
	authorizationv1 "k8s.io/api/authorization/v1"
	authorizationv1beta1 "k8s.io/api/authorization/v1beta1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	autoscalingv2beta1 "k8s.io/api/autoscaling/v2beta1"
	autoscalingv2beta2 "k8s.io/api/autoscaling/v2beta2"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	batchv2alpha1 "k8s.io/api/batch/v2alpha1"
	certificatesv1beta1 "k8s.io/api/certificates/v1beta1"
	coordinationv1 "k8s.io/api/coordination/v1"
	coordinationv1beta1 "k8s.io/api/coordination/v1beta1"
	discoveryv1alpha1 "k8s.io/api/discovery/v1alpha1"
	discoveryv1beta1 "k8s.io/api/discovery/v1beta1"
	eventsv1beta1 "k8s.io/api/events/v1beta1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	flowcontrolv1alpha1 "k8s.io/api/flowcontrol/v1alpha1"
	imagepolicyv1alpha1 "k8s.io/api/imagepolicy/v1alpha1"
	networkingv1 "k8s.io/api/networking/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	nodev1alpha1 "k8s.io/api/node/v1alpha1"
	nodev1beta1 "k8s.io/api/node/v1beta1"
	policyb1beta1 "k8s.io/api/policy/v1beta1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	rbacv1alpha1 "k8s.io/api/rbac/v1alpha1"
	rbacv1beta1 "k8s.io/api/rbac/v1beta1"
	schedulingv1 "k8s.io/api/scheduling/v1"
	schedulingv1alpha1 "k8s.io/api/scheduling/v1alpha1"
	schedulingv1beta1 "k8s.io/api/scheduling/v1beta1"
	settingsv1alpha1 "k8s.io/api/settings/v1alpha1"
	storagev1 "k8s.io/api/storage/v1"
	storagev1alpha1 "k8s.io/api/storage/v1alpha1"
	storagev1beta1 "k8s.io/api/storage/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

	"reflect"
)

var schemes = runtime.NewScheme()

func init() {
	_ = apiextensionsv1.AddToScheme(schemes)
	_ = apiextensionsv1beta1.AddToScheme(schemes)
	_ = autoscalingv1.AddToScheme(schemes)
	_ = coordinationv1.AddToScheme(schemes)
	_ = admissionregistrationv1.AddToScheme(schemes)
	_ = authenticationv1.AddToScheme(schemes)
	_ = policyb1beta1.AddToScheme(schemes)
	_ = networkingv1.AddToScheme(schemes)
	_ = networkingv1beta1.AddToScheme(schemes)
	_ = rbacv1beta1.AddToScheme(schemes)
	_ = appsv1beta2.AddToScheme(schemes)
	_ = admissionv1.AddToScheme(schemes)
	_ = admissionv1beta1.AddToScheme(schemes)
	_ = admissionregistrationv1.AddToScheme(schemes)
	_ = admissionregistrationv1beta1.AddToScheme(schemes)
	_ = appsv1.AddToScheme(schemes)
	_ = appsv1beta1.AddToScheme(schemes)
	_ = appsv1beta2.AddToScheme(schemes)
	_ = auditregistrationv1alpha1.AddToScheme(schemes)
	_ = authenticationv1.AddToScheme(schemes)
	_ = authenticationv1beta1.AddToScheme(schemes)
	_ = authorizationv1.AddToScheme(schemes)
	_ = authorizationv1beta1.AddToScheme(schemes)
	_ = autoscalingv1.AddToScheme(schemes)
	_ = autoscalingv2beta1.AddToScheme(schemes)
	_ = autoscalingv2beta2.AddToScheme(schemes)
	_ = batchv1.AddToScheme(schemes)
	_ = batchv1beta1.AddToScheme(schemes)
	_ = batchv2alpha1.AddToScheme(schemes)
	_ = certificatesv1beta1.AddToScheme(schemes)
	_ = coordinationv1.AddToScheme(schemes)
	_ = coordinationv1beta1.AddToScheme(schemes)
	_ = discoveryv1alpha1.AddToScheme(schemes)
	_ = discoveryv1beta1.AddToScheme(schemes)
	_ = eventsv1beta1.AddToScheme(schemes)
	_ = extensionsv1beta1.AddToScheme(schemes)
	_ = flowcontrolv1alpha1.AddToScheme(schemes)
	_ = imagepolicyv1alpha1.AddToScheme(schemes)
	_ = nodev1alpha1.AddToScheme(schemes)
	_ = nodev1beta1.AddToScheme(schemes)
	_ = policyv1beta1.AddToScheme(schemes)
	_ = rbacv1.AddToScheme(schemes)
	_ = rbacv1alpha1.AddToScheme(schemes)
	_ = rbacv1beta1.AddToScheme(schemes)
	_ = schedulingv1.AddToScheme(schemes)
	_ = schedulingv1alpha1.AddToScheme(schemes)
	_ = schedulingv1beta1.AddToScheme(schemes)
	_ = settingsv1alpha1.AddToScheme(schemes)
	_ = storagev1.AddToScheme(schemes)
	_ = storagev1alpha1.AddToScheme(schemes)
	_ = storagev1beta1.AddToScheme(schemes)
	_ = v1.AddToScheme(schemes)
}

func kubeKinds(prefix string, item *config.Directive) (metav1.Object, bool) {
	userKind, userVersion := utils.Split2(item.Name, ":")
	for reg, regType := range schemes.AllKnownTypes() {
		version, kind := reg.ToAPIVersionAndKind()

		if userVersion == version && userKind == kind {

			objValue := reflect.New(regType)
			obj := objValue.Interface().(metav1.Object)

			typeMeta := objValue.Elem().FieldByName("TypeMeta")
			typeMeta.Set(reflect.ValueOf(metav1.TypeMeta{
				Kind: kind, APIVersion: version,
			}))

			asserts.Metadata(obj, item)
			asserts.AutoLabels(obj, prefix)

			if spec := refs.GetField(obj, "Spec"); spec.IsValid() {
				if spec.Kind() == reflect.Ptr {
					refs.Unmarshal(spec.Interface(), item)
				} else {
					refs.Unmarshal(spec.Addr().Interface(), item)
				}
			} else {
				refs.Unmarshal(obj, item)
			}
			return obj, true
		}
	}
	return nil, false
}
