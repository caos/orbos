// +build test integration

package integration_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/caos/infrop/internal/kinds/clusters/core/infra"
	"github.com/caos/infrop/internal/core/helpers"

	"github.com/caos/infrop/internal/kinds/providers/integration/core"
)

type resourcesNotReturnedError struct {
	name string
}

func (b *resourcesNotReturnedError) Error() string {
	return fmt.Sprintf("Resource %s not returned", b.name)
}

func TestPools(t *testing.T) {
	// TODO: Resolve race conditions
	// t.Parallel()

	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	operatorID := "itpool"
	prov := core.ProvidersUnderTest(configCB)
	if err := core.Cleanup(prov, operatorID); err != nil {
		panic(err)
	}

	args := testPoolArgs(operatorID)
	pools, err := testPools(args)
	if err != nil {
		panic(err)
	}

	for _, pool := range pools {
		var wg sync.WaitGroup
		synchronizer := helpers.NewSynchronizer(&wg)
		wg.Add(1)
		go func() {
			_, addErr := pool.AddCompute()
			synchronizer.Done(addErr)
		}()
		wg.Add(1)
		go func() {
			_, addErr := pool.AddCompute()
			synchronizer.Done(addErr)
		}()
		wg.Wait()

		if synchronizer.IsError() {
			panic(synchronizer)
		}

		computes, err := pool.GetComputes()
		if err != nil {
			panic(err)
		}
		if len(computes) != 2 {
			panic(fmt.Errorf("Expected 2 computes but got %d", len(computes)))
		}
	}

	args.Pools = make([]string, 0)
	pools, err = testPools(args)
	if err != nil {
		panic(err)
	}

	if len(pools) == 0 {
		panic("Not configured pools that still have computes should be returned")
	}

	if err := core.Cleanup(prov, operatorID); err != nil {
		panic(err)
	}
}

func testPoolArgs(operatorID string) core.OperatorArgs {
	return core.OperatorArgs{
		OperatorID:    operatorID,
		Pools:         []string{"unbalancedtestpool"},
		Loadbalancers: make([]*core.LoadBalancer, 0),
	}
}

type ensuredPool struct {
	pool infra.Pool
	err  error
}

func testPools(operatorArgs core.OperatorArgs) ([]infra.Pool, error) {

	prov := core.ProvidersUnderTest(configCB)

	ensured, err := core.Ensure(prov, operatorArgs, nil)
	if err != nil {
		return nil, err
	}

	notReturnedResources := 0
	pools := make([]infra.Pool, 0)

	var wg sync.WaitGroup
	wg.Add(len(ensured))
	synchronizer := helpers.NewSynchronizer(&wg)
	for _, provider := range ensured {
		pool, ok := provider.Pools["unbalancedtestpool"]
		if ok {
			pools = append(pools, pool)
		} else {
			notReturnedResources++
		}
		synchronizer.Done(provider.Err)
	}

	wg.Wait()

	if synchronizer.IsError() {
		if len(pools) > 0 {
			return nil, errors.New("Some providers returned pools, others returned errors")
		}
		return nil, synchronizer
	}

	if notReturnedResources > 0 {
		if notReturnedResources != len(ensured) {
			return nil, errors.New("Some providers returned pools, others did not")
		}
		return nil, &resourcesNotReturnedError{"unbalancedtestpool"}
	}

	return pools, nil
}

func clearPool(pool infra.Pool) error {

	computes, err := pool.GetComputes()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(len(computes))
	synchronizer := helpers.NewSynchronizer(&wg)
	for _, c := range computes {
		go func(compute infra.Compute) {
			synchronizer.Done(compute.Remove())
		}(c)
	}
	wg.Wait()

	if synchronizer.IsError() {
		return synchronizer
	}
	return nil
}
