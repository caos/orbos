package static

import (
	"github.com/caos/orbos/internal/operator/orbiter"
	"github.com/caos/orbos/internal/secret"
	"github.com/caos/orbos/internal/tree"
	"github.com/pkg/errors"
)

type DesiredV0 struct {
	Common *tree.Common `yaml:",inline"`
	//Configruation for the static machines
	Spec Spec
	//Descriptive configuration for the desired loadbalancing to connect the nodes
	Loadbalancing *tree.Tree
}

type Spec struct {
	//Flag to set log-level to debug
	Verbose bool
	//List of Pools with an identification key which will get ensured
	Pools map[string][]*Machine
	//Used SSH-keys used for ensuring
	Keys *Keys
}

type Keys struct {
	//SSH-private-key used for bootstrapping
	BootstrapKeyPrivate *secret.Secret `yaml:",omitempty"`
	//SSH-public-key used for bootstrapping
	BootstrapKeyPublic *secret.Secret `yaml:",omitempty"`
	//SSH-private-key used for maintaining
	MaintenanceKeyPrivate *secret.Secret `yaml:",omitempty"`
	//SSH-public-key used for maintaining
	MaintenanceKeyPublic *secret.Secret `yaml:",omitempty"`
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
	//Used ID for the machine
	ID string
	//IP of the machine to connect to
	IP orbiter.IPAddress
	//Flag if reboot of the machine is required
	RebootRequired *bool
}

func (c *Machine) validate() error {
	if c.ID == "" {
		return errors.New("No id provided")
	}
	return c.IP.Validate()
}
