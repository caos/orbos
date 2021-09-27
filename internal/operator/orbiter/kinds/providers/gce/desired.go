package gce

import (
	"errors"
	"fmt"

	"github.com/caos/orbos/v5/mntr"
	secret2 "github.com/caos/orbos/v5/pkg/secret"
	"github.com/caos/orbos/v5/pkg/tree"
)

type Desired struct {
	Common        *tree.Common `yaml:",inline"`
	Spec          Spec
	Loadbalancing *tree.Tree
}

type Pool struct {
	Multizonal      []string
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
	Private *secret2.Secret `yaml:",omitempty"`
	Public  *secret2.Secret `yaml:",omitempty"`
}

type Spec struct {
	Verbose             bool
	JSONKey             *secret2.Secret `yaml:",omitempty"`
	Region              string
	Zone                string
	Pools               map[string]*Pool
	SSHKey              *SSHKey
	RebootRequired      []string
	ReplacementRequired []string
}

func (d Desired) validateAdapt() (err error) {
	defer func() {
		err = mntr.ToUserError(err)
	}()

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

func (d Desired) validateJSONKey() error {
	if d.Spec.JSONKey == nil || d.Spec.JSONKey.Value == "" {
		return mntr.ToUserError(errors.New("jsonkey missing... please provide a google service accounts jsonkey using orbctl writesecret command"))
	}
	return nil
}

func (d Desired) validateQuery() (err error) {

	defer func() {
		err = mntr.ToUserError(err)
	}()

	if err := d.validateJSONKey(); err != nil {
		return err
	}
	if d.Spec.SSHKey == nil ||
		d.Spec.SSHKey.Private == nil ||
		d.Spec.SSHKey.Private.Value == "" ||
		d.Spec.SSHKey.Public == nil ||
		d.Spec.SSHKey.Public.Value == "" {
		return errors.New("ssh key missing... please initialize your orb using orbctl configure command")
	}
	return nil
}

func parseDesiredV0(desiredTree *tree.Tree) (*Desired, error) {
	desiredKind := &Desired{
		Common: desiredTree.Common,
		Spec:   Spec{},
	}

	if err := desiredTree.Original.Decode(desiredKind); err != nil {
		return nil, mntr.ToUserError(fmt.Errorf("parsing desired state failed: %w", err))
	}

	return desiredKind, nil
}
