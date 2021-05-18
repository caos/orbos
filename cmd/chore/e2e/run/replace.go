package main

import (
	"context"
	"fmt"
	"time"

	"github.com/caos/orbos/internal/operator/common"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
)

var _ testFunc = replace

func replace(settings programSettings, _ *kubernetes.Spec) interactFunc {
	return func(_ uint8, orbctl newOrbctlCommandFunc) (time.Duration, checkCurrentFunc, error) {

		replaceCtx, replaceCancel := context.WithTimeout(settings.ctx, 30*time.Second)
		defer replaceCancel()

		context, nodeID, err := someMasterNodeContextAndID(replaceCtx, settings, orbctl)
		if err != nil {
			return 0, nil, err
		}

		return 15 * time.Minute, func(current common.NodeAgentsCurrentKind) error {
			_, ok := current.Current.Get(nodeID)
			if ok {
				return fmt.Errorf("nodeagent %s still has a current state", nodeID)
			}
			return nil
		}, runCommand(settings, true, nil, nil, orbctl(replaceCtx), "--gitops", "node", "replace", fmt.Sprintf("%s.%s", context, nodeID))
	}
}
