package helm

import (
	"github.com/caos/orbos/pkg/kubernetes/k8s"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func DefaultValues(imageTags map[string]string, image string) *Values {
	return &Values{
		FullnameOverride: "node-exporter",
		Image: &Image{
			Repository: image,
			Tag:        imageTags[image],
			PullPolicy: "IfNotPresent",
		},
		Service: &Service{
			Type:        "ClusterIP",
			Port:        9100,
			TargetPort:  9100,
			NodePort:    "",
			Annotations: map[string]string{"prometheus.io/scrape": "true"},
		},
		Prometheus: &Prometheus{
			Monitor: &Monitor{
				Enabled:          false,
				AdditionalLabels: map[string]string{},
				Namespace:        "",
				ScrapeTimeout:    "10s",
			},
		},
		ServiceAccount: &ServiceAccount{
			Create: true,
		},
		SecurityContext: &SecurityContext{
			RunAsNonRoot: true,
			RunAsUser:    65534,
		},
		Rbac: &Rbac{
			Create:     true,
			PspEnabled: true,
		},
		HostNetwork: false,
		Tolerations: []*Toleration{{
			Effect:   "NoSchedule",
			Operator: "Exists",
		}},
		Resources: &k8s.Resources{
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("50m"),
				corev1.ResourceMemory: resource.MustParse("50Mi"),
			},
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("10m"),
				corev1.ResourceMemory: resource.MustParse("10Mi"),
			},
		},
	}
}
