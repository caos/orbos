package node

import (
	"os"
	"os/exec"

	"github.com/caos/orbiter/internal/core/operator/nodeagent"
	"github.com/caos/orbiter/internal/core/operator/nodeagent/edge/rebooter"
)

func New() nodeagent.Rebooter {
	return rebooter.Func(func() error {
		if err := exec.Command("reboot").Run(); err != nil {
			return err
		}
		os.Exit(0)
		return nil
	})
}
