package helm

import (
	"github.com/caos/orbos/internal/operator/boom/api/latest/k8s"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func DefaultValues(imageTags map[string]string) *Values {
	return &Values{
		FullnameOverride: "kube-state-metrics",
		PrometheusScrape: true,
		Image: &Image{
			Repository: "quay.io/coreos/kube-state-metrics",
			Tag:        imageTags["quay.io/coreos/kube-state-metrics"],
			PullPolicy: "IfNotPresent",
		},
		Replicas: 1,
		Service: &Service{
			Port:           8080,
			Type:           "ClusterIP",
			NodePort:       0,
			LoadBalancerIP: "",
			Annotations:    map[string]string{},
		},
		CustomLabels: map[string]string{},
		HostNetwork:  false,
		Rbac: &Rbac{
			Create: true,
		},
		ServiceAccount: &ServiceAccount{
			Create: true,
			Name:   "",
		},
		Prometheus: &Prometheus{
			Monitor: &Monitor{
				Enabled: false,
			},
		},
		PodSecurityPolicy: &PodSecurityPolicy{
			Enabled: false,
		},
		SecurityContext: &SecurityContext{
			Enabled:   true,
			RunAsUser: 65534,
			FsGroup:   65534,
		},
		NodeSelector:   map[string]string{},
		Affinity:       nil,
		Tolerations:    nil,
		PodAnnotations: map[string]string{},
		Collectors: &Collectors{
			Certificatesigningrequests: true,
			Configmaps:                 true,
			Cronjobs:                   true,
			Daemonsets:                 true,
			Deployments:                true,
			Endpoints:                  true,
			Horizontalpodautoscalers:   true,
			Ingresses:                  true,
			Jobs:                       true,
			Limitranges:                true,
			Namespaces:                 true,
			Nodes:                      true,
			Persistentvolumeclaims:     true,
			Persistentvolumes:          true,
			Poddisruptionbudgets:       true,
			Pods:                       true,
			Replicasets:                true,
			Replicationcontrollers:     true,
			Resourcequotas:             true,
			Secrets:                    true,
			Services:                   true,
			Statefulsets:               true,
			Storageclasses:             true,
			Verticalpodautoscalers:     false,
		},
		Resources: &k8s.Resources{
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("20m"),
				corev1.ResourceMemory: resource.MustParse("100Mi"),
			},
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("10m"),
				corev1.ResourceMemory: resource.MustParse("50Mi"),
			},
		},
	}
}
