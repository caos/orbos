package legacycf

import (
	"fmt"

	"github.com/caos/orbos/v5/mntr"

	"github.com/caos/orbos/v5/internal/operator/networking/kinds/networking/legacycf/config"
	"github.com/caos/orbos/v5/pkg/tree"
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
		return nil, mntr.ToUserError(fmt.Errorf("parsing desired state failed: %w", err))
	}

	return desiredKind, nil
}
