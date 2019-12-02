package noop

import (
	"github.com/caos/orbiter/internal/kinds/nodeagent/adapter"
	"github.com/caos/orbiter/internal/kinds/nodeagent/edge/rebooter"
)

func New() adapter.Rebooter {
	return rebooter.Func(func() error {
		return nil
	})
}
