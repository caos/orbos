package main

import (
	"context"
	"fmt"
	"time"

	"github.com/caos/orbos/internal/operator/common"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
)

var _ testFunc = reboot

func reboot(settings programSettings, _ *kubernetes.Spec) interactFunc {
	return func(_ uint8, orbctl newOrbctlCommandFunc) (time.Duration, checkCurrentFunc, error) {

		rebootCtx, rebootCancel := context.WithTimeout(settings.ctx, 30*time.Second)
		defer rebootCancel()

		var (
			since   = time.Now()
			context = fmt.Sprintf("%s.management", settings.orbID)
			nodeID  string
		)

		if err := runCommand(settings, true, nil, func(line string) {
			nodeID = line
		}, orbctl(rebootCtx), "--gitops", "nodes", "list", "--context", context, "--column", "id"); err != nil {
			return 0, nil, err
		}

		return 10 * time.Minute, func(current common.NodeAgentsCurrentKind) error {
			nodeagent, ok := current.Current.Get(nodeID)
			if !ok {
				return fmt.Errorf("nodeagent %s not found", nodeID)
			}
			if nodeagent.Booted.Before(since) {
				return fmt.Errorf("nodeagent %s has rebooted at %s which is before %s", nodeID, nodeagent.Booted.Format("15:04:05"), since.Format("15:04:05"))
			}
			return nil
		}, runCommand(settings, true, nil, nil, orbctl(rebootCtx), "--gitops", "node", "reboot", fmt.Sprintf("%s.%s", context, nodeID))
	}
}
