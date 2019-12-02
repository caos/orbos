package exit

import (
	"os"

	"github.com/caos/orbiter/internal/kinds/nodeagent/adapter"
	"github.com/caos/orbiter/internal/kinds/nodeagent/edge/rebooter"
)

func New() adapter.Rebooter {
	return rebooter.Func(func() error {
		os.Exit(0)
		return nil
	})
}
