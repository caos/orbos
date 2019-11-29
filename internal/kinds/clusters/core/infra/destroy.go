package infra

import (
	"sync"

	"github.com/caos/orbiter/internal/core/helpers"
)

func Destroy(providers map[string]interface{}) (err error) {

	cleanupping := make([]<-chan error, 0)

	var wg sync.WaitGroup
	synchronizer := helpers.NewSynchronizer(&wg)

	for _, provider := range providers {
		prov, ok := provider.(ProviderCurrent)
		if !ok {
			continue
		}
		pools := prov.Pools()
		cu := prov.Cleanupped()
		cleanupping = append(cleanupping, cu)
		for _, pool := range pools {
			comps, gcErr := pool.GetComputes(true)
			if gcErr != nil {
				err = gcErr
				continue
			}
			for _, comp := range comps {
				wg.Add(1)
				go func(c Compute) {
					if rmErr := c.Remove(); rmErr != nil {
						synchronizer.Done(rmErr)
						return
					}
					synchronizer.Done(nil)
				}(comp)
			}
		}
	}

	for _, cu := range cleanupping {
		if cuErr := <-cu; cuErr != nil {
			err = cuErr
		}
	}

	wg.Wait()

	if synchronizer.IsError() {
		return synchronizer
	}

	return err
}
