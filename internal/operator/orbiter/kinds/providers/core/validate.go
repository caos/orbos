package core

import (
	"fmt"

	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/core/infra"
)

func Validate(lbs map[string]*infra.Ingress, pools []string) error {

	for lbName, lb := range lbs {
		if len(lb.Pools) == 0 {
			return fmt.Errorf("Load balancer %s without pools makes no sense", lbName)
		}
	poolFound:
		for _, targetName := range lb.Pools {
			for _, poolName := range pools {
				if targetName == poolName {
					continue poolFound
				}
			}
			return fmt.Errorf("Load balancing target pool %s is not configured", targetName)
		}
	}

	return nil
}
