package treelabels

import (
	"fmt"
	"strings"

	"github.com/caos/orbos/pkg/labels"
	"github.com/caos/orbos/pkg/tree"
)

func MustForAPI(tree *tree.Tree, operator *labels.Operator) *labels.API {
	if tree == nil ||
		tree.Common == nil ||
		tree.Common.Kind == "" ||
		strings.Count(tree.Common.Kind, "/") != 1 ||
		tree.Common.Version() == "" {
		panic(fmt.Errorf("invalid tree: %+v", tree))
	}

	return labels.MustForAPI(operator, strings.Split(tree.Common.Kind, "/")[1], tree.Common.Version())
}
