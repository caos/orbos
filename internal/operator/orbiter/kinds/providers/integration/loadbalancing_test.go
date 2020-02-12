// +build test integration

package integration_test

import (
	"errors"
	"testing"

	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/integration/core"
)

func TestLoadBalancing(t *testing.T) {
	// TODO: Resolve race conditions
	// t.Parallel()

	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	prov := core.ProvidersUnderTest(configCB)

	operatorID := "itlb"
	if err := core.Cleanup(prov, operatorID); err != nil {
		panic(err)
	}

	args := core.OperatorArgs{
		OperatorID: operatorID,
		Loadbalancers: []*core.LoadBalancer{
			&core.LoadBalancer{
				Name:  "intlb",
				Pools: []string{"balancedmaster", "unbalancedworker"},
			},
			&core.LoadBalancer{
				Name:  "extlb",
				Pools: []string{"balancedmaster"},
			},
		},
	}

	providers, err := core.Ensure(prov, args, nil)
	if err == nil {
		if _, ok := err.(*core.BehaviourDiffersError); ok {
			panic(err)
		}
		panic(errors.New("Not specifying machine resources should return an error"))
	}

	args.Pools = []string{"balancedmaster", "unbalancedworker"}

	providers, err = core.Ensure(prov, args, nil)
	if err != nil {
		panic(err)
	}
	for _, provider := range providers {
		if provider.Err != nil {
			panic(provider.Err)
		}

		_, ok := provider.IPs["intlb"]
		if !ok {
			panic(errors.New("intlb not returned"))
		}

		_, ok = provider.IPs["extlb"]
		if !ok {
			panic(errors.New("intlb not returned"))
		}
	}

	if err := core.Cleanup(prov, operatorID); err != nil {
		panic(err)
	}
}
