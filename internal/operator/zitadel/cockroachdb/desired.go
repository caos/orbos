package cockroachdb

import (
	"github.com/caos/orbos/internal/tree"
)

type DesiredV0 struct {
	Common *tree.Common `yaml:",inline"`
	Spec   Spec
}

type Spec struct {
	Verbose      bool
	ReplicaCount int
}
