// +build test integration

package core

import (
	"sync"

	"github.com/caos/orbos/internal/helpers"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/providers/core"
)

type Provider interface {
	Assemble(operatorID string, configuredPools []string, configuredLoadBalancers []*LoadBalancer) (infra.Provider, core.MachinesService, interface{}, error)
}

type LoadBalancer struct {
	Name  string
	Pools []string
}

type OperatorArgs struct {
	OperatorID    string
	Pools         []string
	Loadbalancers []*LoadBalancer
}
type EnsuredValues struct {
	Pools map[string]infra.Pool
	IPs   map[string]interface{}
	Err   error
}

type BehaviourDiffersError struct {
	msg string
}

func (b *BehaviourDiffersError) Error() string {
	return b.msg
}

func Ensure(testProviders []Provider, operatorArgs OperatorArgs, beforeEnsure func(core.MachinesService) error) ([]*EnsuredValues, error) {

	var mux sync.RWMutex
	var wg sync.WaitGroup
	wg.Add(len(testProviders))
	synchronizer := helpers.NewSynchronizer(&wg)
	provs := make([]infra.Provider, 0)
	for _, testProv := range testProviders {
		go func(testProvider Provider) {
			prov, machinesSvc, _, err := testProvider.Assemble(operatorArgs.OperatorID, operatorArgs.Pools, operatorArgs.Loadbalancers)
			if err != nil {
				synchronizer.Done(err)
				return
			}

			if beforeEnsure != nil {
				if err := beforeEnsure(machinesSvc); err != nil {
					synchronizer.Done(err)
					return
				}
			}

			if prov != nil {
				mux.Lock()
				provs = append(provs, prov)
				mux.Unlock()
			}
			synchronizer.Done(err)
		}(testProv)
	}

	wg.Wait()

	if synchronizer.IsError() {
		if len(provs) > 0 {
			return nil, &BehaviourDiffersError{"Some providers were returned together with some errors"}
		}
		return nil, synchronizer
	}

	ensured := make([]*EnsuredValues, 0)
	for _, prov := range provs {
		wg.Add(1)
		go func(provider infra.Provider) {
			machines, ips, pruned, err := provider.Ensure()
			if err != nil {
				if pruned != nil {
					<-pruned
				}
				synchronizer.Done(err)
				return
			}
			mux.Lock()
			ensured = append(ensured, &EnsuredValues{machines, ips, err})
			mux.Unlock()
			if pruned != nil {
				<-pruned
			}
			synchronizer.Done(nil)
		}(prov)
	}

	wg.Wait()

	if synchronizer.IsError() {
		return nil, synchronizer
	}

	return ensured, nil
}

func Cleanup(testProviders []Provider, operatorID string) error {
	_, err := Ensure(testProviders, OperatorArgs{
		OperatorID:    operatorID,
		Loadbalancers: make([]*LoadBalancer, 0),
		Pools:         make([]string, 0),
	}, func(machines core.MachinesService) error {
		pools, err := machines.ListPools()
		if err != nil {
			return err
		}
		var wg sync.WaitGroup
		synchronizer := helpers.NewSynchronizer(&wg)
		for _, pool := range pools {
			cmps, err := machines.List(pool)
			if err != nil {
				return err
			}
			for _, cmp := range cmps {
				wg.Add(1)
				go func(machine infra.Machine) {
					synchronizer.Done(machine.Remove())
				}(cmp)
			}
		}
		wg.Wait()

		if synchronizer.IsError() {
			return synchronizer
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
