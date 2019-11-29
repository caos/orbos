package concepts

import (
	"github.com/caos/infrop/internal/kinds/providers/core"
	"github.com/caos/infrop/internal/kinds/providers/gce/adapter/resourceservices/forwardingrule"
)

func PlanTCPProxyLoadBalancing(lbName string, resources *core.ResourceFactory, resourcesExecutor *core.Executor, firewallExecutor *core.Executor, services *Services, cfg *Config, ipCB func(lbName string, ip string), groupCB func(poolName string, group core.EnsuredGroup)) error {

	bess, err := backendserviceResources(firewallExecutor, resources, services, cfg, groupCB)
	if err != nil {
		return err
	}

	for _, bes := range bess {
		tp := resources.New(services.TargetProxy, nil, []*core.Resource{bes}, nil)

		fr := resources.New(services.ForwardingRule, &forwardingrule.Config{
			External: true,
			Ports:    cfg.Ports,
		}, []*core.Resource{tp}, func(ensured interface{}) error {
			ipCB(lbName, ensured.(*forwardingrule.Ensured).IP)
			return nil
		})
		if err := resourcesExecutor.Plan(fr); err != nil {
			return err
		}
	}

	return nil
}
