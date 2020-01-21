package kubernetes

import (
	"github.com/caos/orbiter/internal/operator/orbiter"
)

type SecretsV0 struct {
	Common  orbiter.Common `yaml:",inline"`
	Secrets Secrets
}

type Secrets struct {
	Kubeconfig *orbiter.Secret `yaml:",omitempty"`
}
