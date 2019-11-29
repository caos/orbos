package noop

import (
	"github.com/caos/infrop/internal/kinds/nodeagent/model"
	"github.com/caos/infrop/internal/kinds/nodeagent/edge/rebooter"
)

func New() model.Rebooter {
	return rebooter.Func(func() error {
		return nil
	})
}
