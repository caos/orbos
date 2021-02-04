package ingress

import (
	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/kubernetes/resources"
	"github.com/caos/orbos/pkg/labels"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func AdaptFuncToEnsure(namespace string, id labels.IDLabels, host, prefix, service string, servicePort int, annotations map[string]string) (resources.QueryFunc, error) {

	ingress := &v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        id.Name(),
			Namespace:   namespace,
			Labels:      labels.MustK8sMap(id),
			Annotations: annotations,
		},
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{{
				Host: host,
				IngressRuleValue: v1beta1.IngressRuleValue{HTTP: &v1beta1.HTTPIngressRuleValue{Paths: []v1beta1.HTTPIngressPath{{
					Path:     prefix,
					PathType: pathTypePtr(v1beta1.PathTypeImplementationSpecific),
					Backend: v1beta1.IngressBackend{
						ServiceName: service,
						ServicePort: intstr.FromInt(servicePort),
						Resource:    nil,
					},
				}}}},
			}},
		},
	}
	return func(_ kubernetes.ClientInt) (resources.EnsureFunc, error) {
		return func(k8sClient kubernetes.ClientInt) error {
			return k8sClient.ApplyIngress(ingress)
		}, nil
	}, nil
}

func pathTypePtr(pathType v1beta1.PathType) *v1beta1.PathType {
	return &pathType
}

func AdaptFuncToDestroy(namespace, name string) (resources.DestroyFunc, error) {
	return func(k8sClient kubernetes.ClientInt) error {
		return k8sClient.DeleteIngress(namespace, name)
	}, nil
}
