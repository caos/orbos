package kubernetes

import (
	"github.com/caos/orbiter/internal/core/operator/orbiter"
)

type SecretsV0 struct {
	Common  orbiter.Common `yaml:",inline"`
	Secrets struct {
		Kubeconfig orbiter.Secret
	}
	Deps map[string]*orbiter.Tree
}
