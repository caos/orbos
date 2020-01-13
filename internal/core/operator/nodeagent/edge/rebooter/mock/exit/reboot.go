package exit

import (
	"os"

	"github.com/caos/orbiter/internal/core/operator/nodeagent"
	"github.com/caos/orbiter/internal/core/operator/nodeagent/edge/rebooter"
)

func New() nodeagent.Rebooter {
	return rebooter.Func(func() error {
		os.Exit(0)
		return nil
	})
}
