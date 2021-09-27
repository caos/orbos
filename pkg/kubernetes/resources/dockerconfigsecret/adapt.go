package dockerconfigsecret

import (
	kubernetes2 "github.com/caos/orbos/v5/pkg/kubernetes"
	"github.com/caos/orbos/v5/pkg/kubernetes/resources"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func AdaptFuncToEnsure(namespace string, name string, labels map[string]string, data string) (resources.QueryFunc, error) {
	dcs := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Type: corev1.SecretTypeDockerConfigJson,
		StringData: map[string]string{
			corev1.DockerConfigJsonKey: data,
		},
	}

	return func(_ *kubernetes2.Client) (resources.EnsureFunc, error) {
		return func(k8sClient *kubernetes2.Client) error {
			return k8sClient.ApplySecret(dcs)
		}, nil
	}, nil
}

func AdaptFuncToDestroy(namespace, name string) (resources.DestroyFunc, error) {
	return func(client *kubernetes2.Client) error {
		return client.DeleteSecret(namespace, name)
	}, nil
}
