package node

import (
	"os"
	"os/exec"

	"github.com/caos/orbiter/internal/kinds/nodeagent/adapter"
	"github.com/caos/orbiter/internal/kinds/nodeagent/edge/rebooter"
)

func New() adapter.Rebooter {
	return rebooter.Func(func() error {
		if err := exec.Command("reboot").Run(); err != nil {
			return err
		}
		os.Exit(0)
		return nil
	})
}
