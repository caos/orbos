package static

import (
	"github.com/caos/orbiter/internal/core/operator/orbiter"
	"github.com/caos/orbiter/internal/kinds/clusters/core/infra"
)

type DesiredV0 struct {
	Common *orbiter.Common `yaml:",inline"`
	Spec   struct {
		Verbose             bool
		RemoteUser          string
		RemotePublicKeyPath string
		Pools               map[string][]*Compute
		Hoster              string
	}
	Deps *orbiter.Tree
}

type Compute struct {
	ID       string
	Hostname string
	IP       string
}

type Key struct {
	Public  *orbiter.Secret
	Private *orbiter.Secret
}

type SecretsV0 struct {
	Common  *orbiter.Common `yaml:",inline"`
	Deps    *orbiter.Tree
	Secrets struct {
		Bootstrap   Key
		Maintenance Key
	}
}

type Current struct {
	Common  *orbiter.Common `yaml:",inline"`
	Deps    *orbiter.Tree
	Current struct {
		Pools      map[string]infra.Pool
		Ingresses  map[string]infra.Address
		Cleanupped <-chan error `yaml:"-"`
	}
}
