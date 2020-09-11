package certificate

import (
	corev1 "k8s.io/api/core/v1"
)

type secretInternal struct {
	client    string
	name      string
	namespace string
	labels    map[string]string
}

func queryCertificate(desired []*secretInternal, current []corev1.Secret) ([]*secretInternal, []*secretInternal) {
	createSecrets := make([]*secretInternal, 0)
	deleteSecrets := make([]*secretInternal, 0)

	for _, desiredSecret := range desired {
		found := false
		for _, currentSecret := range current {
			if desiredSecret.name == currentSecret.Name && desiredSecret.namespace == currentSecret.Namespace {
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
			if desiredSecret.name == currentSecret.Name && desiredSecret.namespace == currentSecret.Namespace {
				found = true
			}
		}
		if !found {
			deleteSecrets = append(deleteSecrets, &secretInternal{
				name:      currentSecret.Name,
				namespace: currentSecret.Namespace,
				labels:    nil,
			})
		}
	}

	return createSecrets, deleteSecrets
}
