package managed

import (
	"github.com/caos/orbos/internal/tree"
	"github.com/pkg/errors"
)

type DesiredV0 struct {
	Common *tree.Common `yaml:",inline"`
	Spec   Spec
}

type Spec struct {
	Verbose         bool
	ReplicaCount    int               `yaml:"replicaCount,omitempty"`
	StorageCapacity string            `yaml:"storageCapacity,omitempty"`
	StorageClass    string            `yaml:"storageClass,omitempty"`
	NodeSelector    map[string]string `yaml:"nodeSelector,omitempty"`
	Users           []string          `yaml:"users,omitempty"`
}

func parseDesiredV0(desiredTree *tree.Tree) (*DesiredV0, error) {
	desiredKind := &DesiredV0{
		Common: desiredTree.Common,
		Spec:   Spec{},
	}

	if err := desiredTree.Original.Decode(desiredKind); err != nil {
		return nil, errors.Wrap(err, "parsing desired state failed")
	}

	return desiredKind, nil
}
