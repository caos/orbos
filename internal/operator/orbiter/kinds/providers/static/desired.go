package static

import (
	"github.com/caos/orbos/internal/operator/orbiter"
	"github.com/caos/orbos/internal/secret"
	"github.com/caos/orbos/internal/tree"
	"github.com/pkg/errors"
)

type DesiredV0 struct {
	Common        *tree.Common `yaml:",inline"`
	Spec          Spec
	Loadbalancing *tree.Tree
}

type Spec struct {
	Verbose bool
	Pools   map[string][]*Machine
	Keys    *Keys
}

type Keys struct {
	BootstrapKeyPrivate   *secret.Secret `yaml:",omitempty"`
	BootstrapKeyPublic    *secret.Secret `yaml:",omitempty"`
	MaintenanceKeyPrivate *secret.Secret `yaml:",omitempty"`
	MaintenanceKeyPublic  *secret.Secret `yaml:",omitempty"`
}

func (d DesiredV0) validate() error {

	for pool, machines := range d.Spec.Pools {
		for _, machine := range machines {
			if err := machine.validate(); err != nil {
				return errors.Wrapf(err, "Validating machine %s in pool %s failed", machine.ID, pool)
			}
		}
	}
	return nil
}

func parseDesiredV0(desiredTree *tree.Tree) (*DesiredV0, error) {
	desiredKind := &DesiredV0{
		Common: desiredTree.Common,
		Spec:   Spec{},
	}

	if err := desiredTree.Original.Decode(desiredKind); err != nil {
		return nil, errors.Wrap(err, "parsing desired state failed")
	}

	return desiredKind, nil
}

type Machine struct {
	ID             string
	IP             orbiter.IPAddress
	RebootRequired *bool
}

func (c *Machine) validate() error {
	if c.ID == "" {
		return errors.New("No id provided")
	}
	return c.IP.Validate()
}
