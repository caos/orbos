package orb

import (
	"github.com/caos/orbiter/internal/tree"
)


type Current struct {
	Common    *tree.Common `yaml:",inline"`
	Clusters  map[string]*tree.Tree
	Providers map[string]*tree.Tree
}
