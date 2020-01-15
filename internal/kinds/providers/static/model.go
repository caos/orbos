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
	}
	Deps *orbiter.Tree
}

type Compute struct {
	ID       string
	Hostname string
	IP       string
}

type Key struct {
	Public  *orbiter.Secret `yaml:",omitempty"`
	Private *orbiter.Secret `yaml:",omitempty"`
}

type SecretsV0 struct {
	Common  *orbiter.Common `yaml:",inline"`
	Secrets Secrets
}

type Secrets struct {
	Bootstrap   Key
	Maintenance Key
}

type Current struct {
	Common  *orbiter.Common `yaml:",inline"`
	Deps    *orbiter.Tree
	Current struct {
		Pools      map[string]infra.Pool
		Ingresses  map[string]infra.Address
		cleanupped <-chan error `yaml:"-"`
	}
}

func (c *Current) Pools() map[string]infra.Pool {
	return c.Current.Pools
}
func (c *Current) Ingresses() map[string]infra.Address {
	return c.Current.Ingresses
}
func (c *Current) Cleanupped() <-chan error {
	return c.Current.cleanupped
}
