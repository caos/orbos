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

		since := time.Now()

		context, nodeID, err := someMasterNodeContextAndID(rebootCtx, settings, orbctl)
		if err != nil {
			return 0, nil, err
		}

		return 10 * time.Minute, func(current common.NodeAgentsCurrentKind) error {
			nodeagent, ok := current.Current.Get(nodeID)
			if !ok {
				return fmt.Errorf("nodeagent %s not found", nodeID)
			}
			if nodeagent.Booted.Before(since) {
				return fmt.Errorf("nodeagent %s has latest boot at %s which is before %s", nodeID, nodeagent.Booted.Format(time.RFC822), since.Format(time.RFC822))
			}
			return nil
		}, runCommand(settings, true, nil, nil, orbctl(rebootCtx), "--gitops", "node", "reboot", fmt.Sprintf("%s.%s", context, nodeID))
	}
}
