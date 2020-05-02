package infra

import (
	"sync"

	"github.com/caos/orbiter/internal/helpers"
)

func Destroy(providers map[string]interface{}) (err error) {

	var wg sync.WaitGroup
	synchronizer := helpers.NewSynchronizer(&wg)

	for _, provider := range providers {
		prov, ok := provider.(ProviderCurrent)
		if !ok {
			continue
		}
		pools := prov.Pools()
		for _, pool := range pools {
			comps, gcErr := pool.GetMachines()
			if gcErr != nil {
				err = gcErr
				continue
			}
			for _, comp := range comps {
				wg.Add(1)
				go func(c Machine) {
					if rmErr := c.Remove(); rmErr != nil {
						synchronizer.Done(rmErr)
						return
					}
					synchronizer.Done(nil)
				}(comp)
			}
		}
	}

	wg.Wait()

	if synchronizer.IsError() {
		return synchronizer
	}

	return err
}
