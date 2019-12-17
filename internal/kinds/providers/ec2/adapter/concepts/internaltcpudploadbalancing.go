package concepts

import (
	"github.com/caos/orbiter/internal/kinds/providers/core"
	"github.com/caos/orbiter/internal/kinds/providers/gce/adapter/resourceservices/forwardingrule"
)

func PlanInternalTCPUDPLoadBalancing(lbName string, resources *core.ResourceFactory, resourcesExecutor *core.Executor, firewallExecutor *core.Executor, services *Services, cfg *Config, ipCB func(lbName string, ip string), groupCB func(poolName string, group core.EnsuredGroup)) error {

	bes, err := backendserviceResources(firewallExecutor, resources, services, cfg, groupCB)
	if err != nil {
		return err
	}

	fr := resources.New(services.ForwardingRule, &forwardingrule.Config{
		External: false,
		Ports:    cfg.Ports,
	}, []*core.Resource{bes[0]}, func(ensured interface{}) error {
		ipCB(lbName, ensured.(*forwardingrule.Ensured).IP)
		return nil
	})

	return resourcesExecutor.Plan(fr)
}
