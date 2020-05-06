package gce

import (
	"github.com/caos/orbiter/internal/secret"
	"github.com/caos/orbiter/internal/tree"
	"github.com/pkg/errors"
)

type Desired struct {
	Common        *tree.Common `yaml:",inline"`
	Spec          Spec
	Loadbalancing *tree.Tree
}

type Pool struct {
	OSImage     string
	MinCPUCores int
	MinMemoryGB int
	StorageGB   int
}

type SSHKey struct {
	Private *secret.Secret `yaml:",omitempty"`
	Public  *secret.Secret `yaml:",omitempty"`
}

type Spec struct {
	Verbose bool
	JSONKey *secret.Secret `yaml:",omitempty"`
	SSHKey  *SSHKey        `yaml:",omitempty"`
	Region  string
	Zone    string
	Pools   map[string]*Pool
}

func (d Desired) validate() error {
	return nil
}

func parseDesiredV0(desiredTree *tree.Tree, masterkey string) (*Desired, error) {
	desiredKind := &Desired{
		Common: desiredTree.Common,
		Spec: Spec{
			JSONKey: &secret.Secret{Masterkey: masterkey},
			SSHKey: &SSHKey{
				Private: &secret.Secret{Masterkey: masterkey},
				Public:  &secret.Secret{Masterkey: masterkey},
			},
		},
	}

	if err := desiredTree.Original.Decode(desiredKind); err != nil {
		return nil, errors.Wrap(err, "parsing desired state failed")
	}

	return desiredKind, nil
}

func initializeNecessarySecrets(desiredKind *Desired, masterkey string) {
	if desiredKind.Spec.JSONKey == nil {
		desiredKind.Spec.JSONKey = &secret.Secret{Masterkey: masterkey}
	}
	if desiredKind.Spec.SSHKey == nil {
		desiredKind.Spec.SSHKey = &SSHKey{
			Private: &secret.Secret{Masterkey: masterkey},
			Public:  &secret.Secret{Masterkey: masterkey},
		}
	}
}
