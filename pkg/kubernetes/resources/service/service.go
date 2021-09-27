package service

import (
	kubernetes2 "github.com/caos/orbos/v5/pkg/kubernetes"
	"github.com/caos/orbos/v5/pkg/kubernetes/resources"
	"github.com/caos/orbos/v5/pkg/labels"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type Port struct {
	Port       uint16
	Protocol   string
	TargetPort string
	NodePort   uint16
	Name       string
}

func AdaptFuncToEnsure(
	namespace string,
	nameLabels labels.IDLabels,
	ports []Port,
	t string,
	selector *labels.Selector,
	publishNotReadyAddresses bool,
	clusterIP string,
	externalName string,
) (
	resources.QueryFunc,
	error,
) {
	return func(_ kubernetes2.ClientInt) (resources.EnsureFunc, error) {
		portList := make([]corev1.ServicePort, 0)
		for _, port := range ports {
			portList = append(portList, corev1.ServicePort{
				Name:       port.Name,
				Protocol:   corev1.Protocol(port.Protocol),
				Port:       int32(port.Port),
				TargetPort: intstr.Parse(port.TargetPort),
				NodePort:   int32(port.NodePort),
			})
		}
		service := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      nameLabels.Name(),
				Namespace: namespace,
				Labels:    labels.MustK8sMap(nameLabels),
			},
			Spec: corev1.ServiceSpec{
				Ports:                    portList,
				Selector:                 labels.MustK8sMap(selector),
				Type:                     corev1.ServiceType(t),
				PublishNotReadyAddresses: publishNotReadyAddresses,
				ClusterIP:                clusterIP,
				ExternalName:             externalName,
			},
		}

		return func(k8sClient kubernetes2.ClientInt) error {
			return k8sClient.ApplyService(service)
		}, nil
	}, nil
}

func AdaptFuncToDestroy(namespace, name string) (resources.DestroyFunc, error) {
	return func(client kubernetes2.ClientInt) error {
		return client.DeleteService(namespace, name)
	}, nil
}
