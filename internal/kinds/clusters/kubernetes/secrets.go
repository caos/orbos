package kubernetes

import (
	"github.com/caos/orbiter/internal/core/operator/orbiter"
)

type SecretsV0 struct {
	Common  orbiter.Common `yaml:",inline"`
	Secrets Secrets
	Deps    map[string]*orbiter.Tree
}

type Secrets struct {
	Kubeconfig *orbiter.Secret `yaml:",omitempty"`
}
