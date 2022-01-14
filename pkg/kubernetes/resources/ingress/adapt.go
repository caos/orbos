package ingress

import (
	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/kubernetes/resources"
	"github.com/caos/orbos/pkg/labels"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type Arguments struct {
	Namespace   string
	Id          labels.IDLabels
	Host        string
	Prefix      string
	Service     string
	ServicePort uint16
	Annotations map[string]string
}

func AdaptFuncToEnsure(params *Arguments) (resources.QueryFunc, error) {

	ingress := &v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        params.Id.Name(),
			Namespace:   params.Namespace,
			Labels:      labels.MustK8sMap(params.Id),
			Annotations: params.Annotations,
		},
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{{
				Host: params.Host,
				IngressRuleValue: v1beta1.IngressRuleValue{HTTP: &v1beta1.HTTPIngressRuleValue{Paths: []v1beta1.HTTPIngressPath{{
					Path:     params.Prefix,
					PathType: pathTypePtr(v1beta1.PathTypePrefix),
					Backend: v1beta1.IngressBackend{
						ServiceName: params.Service,
						ServicePort: intstr.FromInt(int(params.ServicePort)),
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
