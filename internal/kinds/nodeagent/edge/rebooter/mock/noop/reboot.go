package noop

import (
	"github.com/caos/orbiter/internal/kinds/nodeagent/model"
	"github.com/caos/orbiter/internal/kinds/nodeagent/edge/rebooter"
)

func New() model.Rebooter {
	return rebooter.Func(func() error {
		return nil
	})
}
