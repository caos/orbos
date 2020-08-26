package zitadel

import (
	"github.com/caos/orbos/internal/operator/zitadel/kinds/iam/zitadel/configuration"
	"github.com/caos/orbos/internal/tree"
	"github.com/pkg/errors"
)

type DesiredV0 struct {
	Common     *tree.Common `yaml:",inline"`
	Spec       *Spec
	Database   *tree.Tree `yaml:"database"`
	Networking *tree.Tree `yaml:"networking"`
}

type Spec struct {
	Verbose       bool
	ReplicaCount  int                          `yaml:"replicaCount,omitempty"`
	Configuration *configuration.Configuration `yaml:"configuration"`
	NodeSelector  map[string]string            `yaml:"nodeSelector,omitempty"`
}

func parseDesiredV0(desiredTree *tree.Tree) (*DesiredV0, error) {
	desiredKind := &DesiredV0{
		Common: desiredTree.Common,
		Spec:   &Spec{},
	}

	if err := desiredTree.Original.Decode(desiredKind); err != nil {
		return nil, errors.Wrap(err, "parsing desired state failed")
	}

	return desiredKind, nil
}
