package provided

import (
	"github.com/caos/orbos/internal/tree"
	"github.com/pkg/errors"
)

type DesiredV0 struct {
	Common *tree.Common `yaml:",inline"`
	//Configuration to connect to an existing database
	Spec Spec
}

type Spec struct {
	//Verbose flag to set debug-level to debug
	Verbose bool
	//Namespace where database service exists
	Namespace string
	//URL to connect to database
	URL string
	//Port to connecto to database
	Port string
	//List of users to connect to database
	Users []string
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
