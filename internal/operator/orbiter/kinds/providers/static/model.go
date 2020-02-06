package static

import (
	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/operator/orbiter"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/core/infra"
)

type DesiredV0 struct {
	Common        *orbiter.Common `yaml:",inline"`
	Spec          Spec
	Loadbalancing *orbiter.Tree
}

type Spec struct {
	Verbose             bool
	RemoteUser          string
	RemotePublicKeyPath string
	Pools               map[string][]*Compute
	Keys                Keys
}

type Keys struct {
	BootstrapKeyPrivate   *orbiter.Secret `yaml:",omitempty"`
	BootstrapKeyPublic    *orbiter.Secret `yaml:",omitempty"`
	MaintenanceKeyPrivate *orbiter.Secret `yaml:",omitempty"`
	MaintenanceKeyPublic  *orbiter.Secret `yaml:",omitempty"`
}

func (d DesiredV0) validate() error {
	if d.Spec.RemoteUser == "" {
		return errors.New("No remote user provided")
	}

	if d.Spec.RemotePublicKeyPath == "" {
		return errors.New("No remote public key path provided")
	}

	for pool, computes := range d.Spec.Pools {
		for _, compute := range computes {
			if err := compute.validate(); err != nil {
				return errors.Wrapf(err, "Validating compute %s in pool %s failed", compute.ID, pool)
			}
		}
	}
	return nil
}

type Compute struct {
	ID       string
	Hostname string
	IP       orbiter.IPAddress
}

func (c *Compute) validate() error {
	if c.ID == "" {
		return errors.New("No id provided")
	}
	if c.Hostname == "" {
		return errors.New("No hostname provided")
	}
	return c.IP.Validate()
}

type Current struct {
	Common  *orbiter.Common `yaml:",inline"`
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
