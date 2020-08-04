package pvc

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func AdaptFuncToEnsure(name string, namespace string, labels map[string]string, storageClass string, capacity string) (resources.QueryFunc, error) {

	q, err := resource.ParseQuantity(capacity)
	if err != nil {
		return nil, err
	}
	fs := corev1.PersistentVolumeFilesystem
	pvcDef := &corev1.PersistentVolumeClaim{
		ObjectMeta: v1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
			Labels:    labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &storageClass,
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					"storage": q,
				},
			},
			VolumeMode: &fs,
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
		},
	}

	return func(_ *kubernetes.Client) (resources.EnsureFunc, error) {
		return func(k8sClient *kubernetes.Client) error {
			return k8sClient.ApplyPVC(pvcDef)
		}, nil
	}, nil
}

func AdaptFuncToDestroy(name, namespace string) (resources.DestroyFunc, error) {
	return func(k8sClient *kubernetes.Client) error {
		//TODO
		return nil
	}, nil
}
