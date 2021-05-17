package helm

import (
	"github.com/caos/orbos/pkg/kubernetes/k8s"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func DefaultValues(imageTags map[string]string, image string) *Values {
	return &Values{
		Operator: &Operator{
			ClusterDomain: "",
			NsToWatch:     "",
			Image: &Image{
				Repository: image,
				Tag:        imageTags[image],
				PullPolicy: "IfNotPresent",
			},
			ImagePullSecrets: []interface{}{},
			ReplicaCount:     1,
			SecurityContext: &SecurityContext{
				RunAsUser:    1000,
				RunAsGroup:   1000,
				RunAsNonRoot: true,
				FsGroup:      1000,
			},
			Resources: &k8s.Resources{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:              resource.MustParse("200m"),
					corev1.ResourceMemory:           resource.MustParse("256Mi"),
					corev1.ResourceEphemeralStorage: resource.MustParse("500Mi"),
				},
			},
		},
		Console: &Console{
			Image: &Image{
				Repository: "minio/console",
				Tag:        imageTags["minio/console"],
				PullPolicy: "IfNotPresent",
			},
			ReplicaCount: 1,
			Resources:    &k8s.Resources{},
		},
		Tenants: []*Tenants{},
	}
}
