package certificates

import (
	corev1 "k8s.io/api/core/v1"
)

type SecretInternal struct {
	Client    string
	Name      string
	Namespace string
	Labels    map[string]string
}

func QueryCertificate(desired []*SecretInternal, current []corev1.Secret) ([]*SecretInternal, []*SecretInternal) {
	createSecrets := make([]*SecretInternal, 0)
	deleteSecrets := make([]*SecretInternal, 0)

	for _, desiredSecret := range desired {
		found := false
		for _, currentSecret := range current {
			if desiredSecret.Name == currentSecret.Name && desiredSecret.Namespace == currentSecret.Namespace {
				found = true
			}
		}
		if !found {
			createSecrets = append(createSecrets, desiredSecret)
		}
	}

	for _, currentSecret := range current {
		found := false
		for _, desiredSecret := range desired {
			if desiredSecret.Name == currentSecret.Name && desiredSecret.Namespace == currentSecret.Namespace {
				found = true
			}
		}
		if !found {
			deleteSecrets = append(deleteSecrets, &SecretInternal{
				Name:      currentSecret.Name,
				Namespace: currentSecret.Namespace,
				Labels:    nil,
			})
		}
	}

	return createSecrets, deleteSecrets
}
