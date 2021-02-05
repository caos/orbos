package legacycf

import (
	"github.com/caos/orbos/internal/operator/networking/kinds/networking/legacycf/config"
	"github.com/caos/orbos/pkg/tree"
	"github.com/pkg/errors"
)

type Desired struct {
	Common *tree.Common           `yaml:",inline"`
	Spec   *config.ExternalConfig `yaml:"spec"`
}

func parseDesired(desiredTree *tree.Tree) (*Desired, error) {
	desiredKind := &Desired{
		Common: desiredTree.Common,
		Spec:   &config.ExternalConfig{},
	}

	if err := desiredTree.Original.Decode(desiredKind); err != nil {
		return nil, errors.Wrap(err, "parsing desired state failed")
	}

	return desiredKind, nil
}
