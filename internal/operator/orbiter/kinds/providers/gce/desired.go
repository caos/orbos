package gce

import (
	"fmt"

	"github.com/caos/orbos/internal/secret"
	"github.com/caos/orbos/internal/tree"
	"github.com/pkg/errors"
)

type Desired struct {
	Common *tree.Common `yaml:",inline"`
	//Configruation for the virtual machines on GCE
	Spec Spec
	//Descriptive configuration for the desired loadbalancing to connect the nodes
	Loadbalancing *tree.Tree
}

type Pool struct {
	//Used OS-image for the VMs in the pool
	OSImage string `yaml:"osimage"`
	//Minimum of requested v-CPU-cores for the VMs in the pool
	MinCPUCores int `yaml:"mincpucores"`
	//Minimum of requested memory for the VMs in the pool
	MinMemoryGB int `yaml:"minmemorygb"`
	//GB of storage requestes for the VMs in the pool
	StorageGB int `yaml:"storagegb"`
	//Flag if VMs should be preemptible and can be shutdown and restarted after 24h
	Preemptible bool
	//Count of mounted local SSDs with a size of 370 GB
	LocalSSDs uint8 `yaml:"localssds"`
}

func (p Pool) validate() error {

	if p.MinCPUCores == 0 {
		return errors.New("no cpu cores configured")
	}
	if p.MinMemoryGB == 0 {
		return errors.New("no memory configured")
	}
	if p.StorageGB == 0 {
		return errors.New("no storage configured")
	}
	switch p.OSImage {
	case
		"projects/gce-uefi-images/global/images/centos-7-v20200403",
		"projects/centos-cloud/global/images/centos-7-v20200429":
		if p.StorageGB < 20 {
			return fmt.Errorf("at least 20GB of storage is needed for image %s", p.OSImage)
		}
	default:
		return fmt.Errorf("OSImage \"%s\" is not supported", p.OSImage)
	}
	return nil
}

type SSHKey struct {
	//Private-SSH-key for ssh-connection to the VMs on GCE
	Private *secret.Secret `yaml:",omitempty"`
	//Public-SSH-key for ssh-connection to the VMs on GCE
	Public *secret.Secret `yaml:",omitempty"`
}

type Spec struct {
	//Flag to set log-level to debug
	Verbose bool
	//Service account key used to create and maintain all elements on GCE
	JSONKey *secret.Secret `yaml:"jsonkey,omitempty"`
	//Region used for all elements on GCE which are region specific
	Region string
	//Zone used for all elements on GCE which are zone specific
	Zone string
	//List of Pools with an identification key which will get ensured
	Pools map[string]*Pool
	//SSH-key for connection to the VMs on GCE
	SSHKey *SSHKey `yaml:"sshkey"`
	//List of nodes which are required to reboot
	RebootRequired []string
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
