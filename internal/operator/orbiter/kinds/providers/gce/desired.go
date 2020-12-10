package gce

import (
	"fmt"

	"github.com/caos/orbos/internal/secret"
	"github.com/caos/orbos/internal/tree"
	"github.com/pkg/errors"
)

type Desired struct {
	Common        *tree.Common `yaml:",inline"`
	Spec          Spec
	Loadbalancing *tree.Tree
}

type Pool struct {
	OSImage         string
	MinCPUCores     int
	MinMemoryGB     int
	StorageGB       int
	StorageDiskType string
	Preemptible     bool
	LocalSSDs       uint8
}

func (p Pool) validate() error {

	if p.MinCPUCores == 0 {
		return errors.New("no cpu cores configured")
	}
	if p.MinMemoryGB == 0 {
		return errors.New("no memory configured")
	}
	if p.StorageGB < 20 {
		return fmt.Errorf("at least 20GB of storage is needed for the boot disk")
	}

	switch p.StorageDiskType {
	case "pd-standard",
		"pd-balanced",
		"pd-ssd":
		break
	default:
		return fmt.Errorf("DiskType \"%s\" is not supported", p.StorageDiskType)
	}

	return nil
}

type SSHKey struct {
	Private *secret.Secret `yaml:",omitempty"`
	Public  *secret.Secret `yaml:",omitempty"`
}

type Spec struct {
	Verbose             bool
	JSONKey             *secret.Secret `yaml:",omitempty"`
	Region              string
	Zone                string
	Pools               map[string]*Pool
	SSHKey              *SSHKey
	RebootRequired      []string
	ReplacementRequired []string
}

func (d Desired) validate() error {
	if d.Loadbalancing == nil {
		return errors.New("no loadbalancing configured")
	}
	if d.Spec.Region == "" {
		return errors.New("no region configured")
	}
	if d.Spec.Zone == "" {
		return errors.New("no zone configured")
	}
	if len(d.Spec.Pools) == 0 {
		return errors.New("no pools configured")
	}
	for poolName, pool := range d.Spec.Pools {
		if err := pool.validate(); err != nil {
			return fmt.Errorf("configuring pool %s failed: %w", poolName, err)
		}
	}
	return nil
}
func parseDesiredV0(desiredTree *tree.Tree) (*Desired, error) {
	desiredKind := &Desired{
		Common: desiredTree.Common,
		Spec:   Spec{},
	}

	if err := desiredTree.Original.Decode(desiredKind); err != nil {
		return nil, errors.Wrap(err, "parsing desired state failed")
	}

	return desiredKind, nil
}
