package static

import (
	"fmt"
	"regexp"

	"github.com/caos/orbos/internal/operator/orbiter"
	secret2 "github.com/caos/orbos/pkg/secret"
	"github.com/caos/orbos/pkg/tree"
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
	Keys               *Keys
	ExternalInterfaces []string
}

type Keys struct {
	//SSH-private-key used for bootstrapping
	BootstrapKeyPrivate *secret2.Secret `yaml:",omitempty"`
	//SSH-public-key used for bootstrapping
	BootstrapKeyPublic *secret2.Secret `yaml:",omitempty"`
	//SSH-private-key used for maintaining
	MaintenanceKeyPrivate *secret2.Secret `yaml:",omitempty"`
	//SSH-public-key used for maintaining
	MaintenanceKeyPublic *secret2.Secret `yaml:",omitempty"`
}

func (d DesiredV0) validateAdapt() error {

	for pool, machines := range d.Spec.Pools {
		for _, machine := range machines {
			if err := machine.validate(); err != nil {
				return errors.Wrapf(err, "Validating machine %s in pool %s failed", machine.ID, pool)
			}
		}
	}
	return nil
}

func (d DesiredV0) validateQuery() error {

	if d.Spec.Keys == nil ||
		d.Spec.Keys.BootstrapKeyPrivate == nil ||
		d.Spec.Keys.BootstrapKeyPrivate.Value == "" {
		return errors.New("bootstrap private ssh key missing... please provide a private ssh bootstrap key using orbctl writesecret command")
	}

	if d.Spec.Keys.MaintenanceKeyPrivate == nil ||
		d.Spec.Keys.MaintenanceKeyPrivate.Value == "" ||
		d.Spec.Keys.MaintenanceKeyPublic == nil ||
		d.Spec.Keys.MaintenanceKeyPublic.Value == "" {
		return errors.New("maintenance ssh key missing... please initialize your orb using orbctl configure command")
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
	//Used hostname for the machine
	Hostname string
	//IP of the machine to connect to
	IP orbiter.IPAddress
	//Flag if reboot of the machine is required
	RebootRequired bool
	//Flag if reqplce of the machine is required
	ReplacementRequired bool
}

var internetHosts = regexp.MustCompile("^[a-z0-9]([-a-z0-9]*[a-z0-9])?$")

func validateName(name string) error {
	if len(name) > 63 || !internetHosts.MatchString(name) {
		return errors.Errorf("name must be compatible with https://tools.ietf.org/html/rfc1123#section-2, but %s is not", name)
	}
	return nil
}

func (c *Machine) validate() error {

	if err := validateName(c.ID); err != nil {
		return fmt.Errorf("validating id failed: %w", err)
	}

	if err := validateName(c.Hostname); err != nil {
		return fmt.Errorf("validating hostname failed: %w", err)
	}

	return c.IP.Validate()
}
