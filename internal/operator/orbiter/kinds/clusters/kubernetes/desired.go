package kubernetes

import (
	"fmt"

	secret2 "github.com/caos/orbos/pkg/secret"

	"github.com/caos/orbos/pkg/tree"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"

	"github.com/caos/orbos/internal/operator/orbiter"
)

var ipPartRegex = `([0-9]|[1-9][0-9]|1[0-9][0-9]|2[0-4][0-9]|25[0-5])`

var ipRegex = fmt.Sprintf(`%s\.%s\.%s\.%s`, ipPartRegex, ipPartRegex, ipPartRegex, ipPartRegex)

var cidrRegex = fmt.Sprintf(`%s/([1-2][0-9]|3[0-2]|[0-9])`, ipRegex)

type DesiredV0 struct {
	Common tree.Common `yaml:",inline"`
	//Configuration for Kubernetes
	Spec Spec
}

type Spec struct {
	//Configuration for the control plane for Kubernetes
	ControlPlane Pool
	//Admin-kubeconfig
	Kubeconfig *secret2.Secret `yaml:",omitempty"`
	//Configuration for the networking in kubernetes
	Networking Networking
	//Flag to set log-level to debug
	Verbose bool
	//Versions to ensure for the components
	Versions CompVersions
	// Use this registry to pull all kubernetes and ORBITER container images from
	//@default: ghcr.io
	CustomImageRegistry string
	//List of configurations for the worker pools
	Workers []*Pool
}

type Networking struct {
	//Used Domain for the kube-dns
	DNSDomain string
	//Used networking solution in kubernetes
	Network string
	//CIDR used for services
	ServiceCidr orbiter.CIDR
	//CIDR used for pods
	PodCidr orbiter.CIDR
}

type CompVersions struct {
	//Ensured version ofKubernetes
	Kubernetes string
	//Self-reconciling version of the Orbiter
	Orbiter string
}

func parseDesiredV0(desiredTree *tree.Tree) (*DesiredV0, error) {
	desiredKind := &DesiredV0{
		Common: *desiredTree.Common,
		Spec:   Spec{},
	}
	if err := desiredTree.Original.Decode(desiredKind); err != nil {
		return nil, errors.Wrap(err, "parsing desired state failed")
	}

	return desiredKind, nil
}

func (d *DesiredV0) validate() error {

	if d.Spec.ControlPlane.Nodes != 1 && d.Spec.ControlPlane.Nodes != 3 && d.Spec.ControlPlane.Nodes != 5 {
		return errors.Errorf("Controlplane nodes can only be scaled to 1, 3 or 5 but desired are %d", d.Spec.ControlPlane.Nodes)
	}

	if ParseString(d.Spec.Versions.Kubernetes) == Unknown {
		return errors.Errorf("Unknown kubernetes version %s", d.Spec.Versions.Kubernetes)
	}

	if err := d.Spec.Networking.ServiceCidr.Validate(); err != nil {
		return err
	}

	if err := d.Spec.Networking.PodCidr.Validate(); err != nil {
		return err
	}

	seenPools := map[string][]string{
		d.Spec.ControlPlane.Provider: {d.Spec.ControlPlane.Pool},
	}

	for _, worker := range d.Spec.Workers {
		pools, ok := seenPools[worker.Provider]
		if !ok {
			seenPools[worker.Provider] = []string{worker.Pool}
			continue
		}
		for _, seenPool := range pools {
			if seenPool == worker.Pool {
				return errors.Errorf("Pool %s from provider %s is used multiple times", worker.Pool, worker.Provider)
			}
		}
		seenPools[worker.Provider] = append(pools, worker.Pool)
	}

	return nil
}

type Pool struct {
	//Flag to disable updates on the nodes of the pool
	UpdatesDisabled bool
	//Provider which should be used to ensure this pool
	Provider string
	//Count of nodes
	Nodes int
	//Name of the pool
	Pool string
	//Taints for the nodes in this pool
	Taints *Taints `yaml:"taints,omitempty"`
}

type Taint struct {
	//Key of the taint
	Key string `yaml:"key"`
	//Value of the taint
	Value string `yaml:"value,omitempty"`
	//Effect of the taint
	Effect core.TaintEffect `yaml:"effect"`
}

type Taints []Taint

func (t *Taints) ToK8sTaints() []core.Taint {
	if t == nil {
		return nil
	}
	taints := make([]core.Taint, len(*t))
	for idx, taint := range *t {
		taints[idx] = core.Taint{
			Key:    taint.Key,
			Value:  taint.Value,
			Effect: taint.Effect,
		}
	}
	return taints
}
