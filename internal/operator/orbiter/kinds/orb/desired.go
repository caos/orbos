package orb

import (
	"errors"
	"fmt"

	"github.com/caos/orbos/v5/mntr"
	"github.com/caos/orbos/v5/pkg/tree"
)

type DesiredV0 struct {
	Common *tree.Common `yaml:",inline"`
	Spec   struct {
		Verbose bool
		PProf   bool
	}
	Clusters  map[string]*tree.Tree
	Providers map[string]*tree.Tree
}

func ParseDesiredV0(desiredTree *tree.Tree) (*DesiredV0, error) {
	desiredKind := &DesiredV0{Common: desiredTree.Common}

	if err := desiredTree.Original.Decode(desiredKind); err != nil {
		return nil, mntr.ToUserError(fmt.Errorf("parsing desired state failed: %w", err))
	}
	desiredKind.Common.OverwriteVersion("v0")

	return desiredKind, nil
}

func (d *DesiredV0) validate() (err error) {
	defer func() {
		err = mntr.ToUserError(err)
	}()
	if len(d.Clusters) < 1 {
		return errors.New("no clusters configured")
	}
	if len(d.Providers) < 1 {
		return errors.New("no providers configured")
	}

	k8sKind := "orbiter.caos.ch/KubernetesCluster"
	var k8s int
	for _, cluster := range d.Clusters {
		if cluster.Common.Kind == k8sKind {
			k8s++
		}
	}
	if k8s != 1 {
		return fmt.Errorf("exactly one cluster of kind %s must be configured, but got %d", k8sKind, k8s)
	}
	return nil
}
