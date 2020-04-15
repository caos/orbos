package orb

import (
	"github.com/caos/orbiter/internal/tree"
	"github.com/pkg/errors"
)

type DesiredV0 struct {
	Common *tree.Common `yaml:",inline"`
	Spec   struct {
		Verbose bool
	}
	Clusters  map[string]*tree.Tree
	Providers map[string]*tree.Tree
}

func parseDesiredV0(desiredTree *tree.Tree) (*DesiredV0, error) {
	desiredKind := &DesiredV0{Common: desiredTree.Common}

	if err := desiredTree.Original.Decode(desiredKind); err != nil {
		return nil, errors.Wrap(err, "parsing desired state failed")
	}
	desiredKind.Common.Version = "v0"

	return desiredKind, nil
}

func (d *DesiredV0) validate() error {
	if len(d.Clusters) < 1 {
		return errors.New("No clusters configured")
	}
	if len(d.Providers) < 1 {
		return errors.New("No providers configured")
	}

	k8sKind := "orbiter.caos.ch/KubernetesCluster"
	var k8s int
	for _, cluster := range d.Clusters {
		if cluster.Common.Kind == k8sKind {
			k8s++
		}
	}
	if k8s != 1 {
		return errors.Errorf("Exactly one cluster of kind %s must be configured, but got %d", k8sKind, k8s)
	}
	return nil
}
