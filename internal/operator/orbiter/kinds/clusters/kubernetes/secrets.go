package kubernetes

import (
	"github.com/caos/orbos/v5/pkg/secret"
)

func getSecretsMap(desiredKind *DesiredV0) map[string]*secret.Secret {
	if desiredKind.Spec.Kubeconfig == nil {
		desiredKind.Spec.Kubeconfig = &secret.Secret{}
	}
	return map[string]*secret.Secret{
		"kubeconfig": desiredKind.Spec.Kubeconfig,
	}
}
