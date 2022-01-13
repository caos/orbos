package static

import (
	"errors"
	"fmt"
	"net"
	"regexp"

	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/secret"
	"github.com/caos/orbos/pkg/tree"
)

type DesiredV0 struct {
	Common        *tree.Common `yaml:",inline"`
	Spec          Spec
	Loadbalancing *tree.Tree
}

type Spec struct {
	Verbose            bool
	Pools              map[string][]*Machine
	Keys               *Keys
	ExternalInterfaces []string
	PrivateInterface   string
}

type Keys struct {
	BootstrapKeyPrivate   *secret.Secret `yaml:",omitempty"`
	BootstrapKeyPublic    *secret.Secret `yaml:",omitempty"`
	MaintenanceKeyPrivate *secret.Secret `yaml:",omitempty"`
	MaintenanceKeyPublic  *secret.Secret `yaml:",omitempty"`
}

func (d DesiredV0) validateAdapt() (err error) {
	defer func() {
		err = mntr.ToUserError(err)
	}()

	for pool, machines := range d.Spec.Pools {
		for _, machine := range machines {
			if err := machine.validate(); err != nil {
				return fmt.Errorf("validating machine %s in pool %s failed: %w", machine.ID, pool, err)
			}
		}
	}
	return nil
}

func (d DesiredV0) validateQuery() (err error) {
	defer func() {
		err = mntr.ToUserError(err)
	}()

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

func parseDesiredV0(desiredTree *tree.Tree) (desiredKind *DesiredV0, err error) {
	defer func() {
		err = mntr.ToUserError(err)
	}()

	desiredKind = &DesiredV0{
		Common: desiredTree.Common,
		Spec:   Spec{},
	}

	if err := desiredTree.Original.Decode(desiredKind); err != nil {
		return nil, fmt.Errorf("parsing desired state failed: %w", err)
	}

	return desiredKind, nil
}

type Machine struct {
	ID                  string
	Hostname            string
	IP                  string
	RebootRequired      bool
	ReplacementRequired bool
}

var internetHosts = regexp.MustCompile("^[a-z0-9]([-a-z0-9]*[a-z0-9])?$")

func validateName(name string) (err error) {
	defer func() {
		err = mntr.ToUserError(err)
	}()

	if len(name) > 63 || !internetHosts.MatchString(name) {
		return fmt.Errorf("name must be compatible with https://tools.ietf.org/html/rfc1123#section-2, but %s is not", name)
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

	if net.ParseIP(c.IP) == nil {
		return fmt.Errorf("%s is not a valid ip address", c.IP)
	}
	return nil
}
