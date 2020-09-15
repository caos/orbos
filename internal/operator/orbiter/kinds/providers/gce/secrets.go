package gce

import (
	"github.com/caos/orbos/internal/secret"
)

func getSecretsMap(desiredKind *Desired) map[string]*secret.Secret {
	if desiredKind.Spec.JSONKey == nil {
		desiredKind.Spec.JSONKey = &secret.Secret{}
	}

	return map[string]*secret.Secret{
		"jsonkey": desiredKind.Spec.JSONKey,
	}
}
