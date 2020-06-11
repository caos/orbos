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
	Verbose             bool
	RemoteUser          string
	RemotePublicKeyPath string
	Pools               map[string][]*Machine
	Keys                Keys
}

type Keys struct {
	BootstrapKeyPrivate   *secret.Secret `yaml:",omitempty"`
	BootstrapKeyPublic    *secret.Secret `yaml:",omitempty"`
	MaintenanceKeyPrivate *secret.Secret `yaml:",omitempty"`
	MaintenanceKeyPublic  *secret.Secret `yaml:",omitempty"`
}

func (k *Keys) MarshalYAML() (interface{}, error) {
	type Alias Keys
	return &Alias{
		BootstrapKeyPrivate:   secret.ClearEmpty(k.BootstrapKeyPrivate),
		BootstrapKeyPublic:    secret.ClearEmpty(k.BootstrapKeyPublic),
		MaintenanceKeyPrivate: secret.ClearEmpty(k.MaintenanceKeyPrivate),
		MaintenanceKeyPublic:  secret.ClearEmpty(k.MaintenanceKeyPublic),
	}, nil
}

func (d DesiredV0) validate() error {
	if d.Spec.RemoteUser == "" {
		return errors.New("No remote user provided")
	}

	if d.Spec.RemotePublicKeyPath == "" {
		return errors.New("No remote public key path provided")
	}

	for pool, machines := range d.Spec.Pools {
		for _, machine := range machines {
			if err := machine.validate(); err != nil {
				return errors.Wrapf(err, "Validating machine %s in pool %s failed", machine.ID, pool)
			}
		}
	}
	return nil
}

func parseDesiredV0(desiredTree *tree.Tree, masterkey string) (*DesiredV0, error) {
	desiredKind := &DesiredV0{
		Common: desiredTree.Common,
		Spec: Spec{
			Keys: Keys{
				BootstrapKeyPrivate:   &secret.Secret{Masterkey: masterkey},
				BootstrapKeyPublic:    &secret.Secret{Masterkey: masterkey},
				MaintenanceKeyPrivate: &secret.Secret{Masterkey: masterkey},
				MaintenanceKeyPublic:  &secret.Secret{Masterkey: masterkey},
			},
		},
	}

	if err := desiredTree.Original.Decode(desiredKind); err != nil {
		return nil, errors.Wrap(err, "parsing desired state failed")
	}

	return desiredKind, nil
}

func rewriteMasterkeyDesiredV0(old *DesiredV0, masterkey string) *DesiredV0 {
	if old != nil {
		newD := new(DesiredV0)
		*newD = *old

		if newD.Spec.Keys.BootstrapKeyPrivate != nil {
			newD.Spec.Keys.BootstrapKeyPrivate.Masterkey = masterkey
		}
		if newD.Spec.Keys.BootstrapKeyPublic != nil {
			newD.Spec.Keys.BootstrapKeyPublic.Masterkey = masterkey
		}
		if newD.Spec.Keys.MaintenanceKeyPrivate != nil {
			newD.Spec.Keys.MaintenanceKeyPrivate.Masterkey = masterkey
		}
		if newD.Spec.Keys.MaintenanceKeyPublic != nil {
			newD.Spec.Keys.MaintenanceKeyPublic.Masterkey = masterkey
		}
		return newD
	}
	return old
}

func initializeNecessarySecrets(desiredKind *DesiredV0, masterkey string) {
	if desiredKind.Spec.Keys.BootstrapKeyPrivate == nil {
		desiredKind.Spec.Keys.BootstrapKeyPrivate = &secret.Secret{Masterkey: masterkey}
	}

	if desiredKind.Spec.Keys.BootstrapKeyPublic == nil {
		desiredKind.Spec.Keys.BootstrapKeyPublic = &secret.Secret{Masterkey: masterkey}
	}

	if desiredKind.Spec.Keys.MaintenanceKeyPrivate == nil {
		desiredKind.Spec.Keys.MaintenanceKeyPrivate = &secret.Secret{Masterkey: masterkey}
	}

	if desiredKind.Spec.Keys.MaintenanceKeyPublic == nil {
		desiredKind.Spec.Keys.MaintenanceKeyPublic = &secret.Secret{Masterkey: masterkey}
	}
}

type Machine struct {
	ID       string
	Hostname string
	IP       orbiter.IPAddress
}

func (c *Machine) validate() error {
	if c.ID == "" {
		return errors.New("No id provided")
	}
	if c.Hostname == "" {
		return errors.New("No hostname provided")
	}
	return c.IP.Validate()
}
