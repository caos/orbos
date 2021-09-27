package kubernetes

import (
	"fmt"

	core "k8s.io/api/core/v1"

	"github.com/caos/orbos/v5/internal/operator/orbiter"
	"github.com/caos/orbos/v5/mntr"
	"github.com/caos/orbos/v5/pkg/secret"
	"github.com/caos/orbos/v5/pkg/tree"
)

type DesiredV0 struct {
	Common tree.Common `yaml:",inline"`
	Spec   Spec
}

type Spec struct {
	ControlPlane Pool
	Kubeconfig   *secret.Secret `yaml:",omitempty"`
	Networking   struct {
		DNSDomain   string
		Network     string
		ServiceCidr orbiter.CIDR
		PodCidr     orbiter.CIDR
	}
	Verbose  bool
	Versions struct {
		Kubernetes string
		Orbiter    string
	}
	// Use this registry to pull all kubernetes and ORBITER container images from
	//@default: ghcr.io
	CustomImageRegistry string
	Workers             []*Pool
}

func parseDesiredV0(desiredTree *tree.Tree) (*DesiredV0, error) {
	desiredKind := &DesiredV0{
		Common: *desiredTree.Common,
		Spec:   Spec{},
	}
	if err := desiredTree.Original.Decode(desiredKind); err != nil {
		return nil, mntr.ToUserError(fmt.Errorf("parsing desired state failed: %w", err))
	}

	return desiredKind, nil
}

func (d *DesiredV0) validate() (err error) {

	defer func() {
		err = mntr.ToUserError(err)
	}()

	if d.Spec.ControlPlane.Nodes != 1 && d.Spec.ControlPlane.Nodes != 3 && d.Spec.ControlPlane.Nodes != 5 {
		return fmt.Errorf("controlplane nodes can only be scaled to 1, 3 or 5 but desired are %d", d.Spec.ControlPlane.Nodes)
	}

	if ParseString(d.Spec.Versions.Kubernetes) == Unknown {
		return fmt.Errorf("unknown kubernetes version %s", d.Spec.Versions.Kubernetes)
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
				return fmt.Errorf("pool %s from provider %s is used multiple times", worker.Pool, worker.Provider)
			}
		}
		seenPools[worker.Provider] = append(pools, worker.Pool)
	}

	return nil
}

type Pool struct {
	UpdatesDisabled bool
	Provider        string
	Nodes           int
	Pool            string
	Taints          *Taints `yaml:"taints,omitempty"`
}

type Taint struct {
	Key    string           `yaml:"key"`
	Value  string           `yaml:"value,omitempty"`
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
