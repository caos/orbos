package concepts

import (
	"strconv"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/providers/core"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/providers/gce/adapter/resourceservices/backendservice"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/providers/gce/adapter/resourceservices/firewallrule"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/providers/gce/adapter/resourceservices/healthcheck"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/providers/gce/adapter/resourceservices/instancegroup"
)

type Services struct {
	BackendService core.ResourceService
	InstanceGroup  core.ResourceService
	HealthChecks   core.ResourceService
	ForwardingRule core.ResourceService
	FirewallRule   core.ResourceService
	TargetProxy    core.ResourceService
}

type Ensured struct {
	Name string
	IP   string
}

type Config struct {
	HealthChecks *healthcheck.Config
	Pools        []string
	Ports        []int64
	External     bool
}

func backendserviceResources(firewallExecutor *core.Executor, resources *core.ResourceFactory, services *Services, cfg *Config, groupCB func(poolName string, group core.EnsuredGroup)) ([]*core.Resource, error) {

	// TODO: Granularize
	allowPorts := []string{
		strconv.FormatInt(cfg.HealthChecks.Port, 10),
	}
	for _, port := range cfg.Ports {
		allowPorts = append(allowPorts, strconv.FormatInt(port, 10))
	}

	fwr := resources.New(services.FirewallRule, &firewallrule.Config{
		Egress:       false,
		IPRanges:     []string{"130.211.0.0/22", "35.191.0.0/16"},
		AllowedPorts: allowPorts,
		DeniedPorts:  nil,
	}, nil, nil)
	if err := firewallExecutor.Plan(fwr); err != nil {
		return nil, err
	}

	besDeps := []*core.Resource{
		resources.New(services.HealthChecks, cfg.HealthChecks, nil, nil),
	}

	for _, pool := range cfg.Pools {
		besDeps = append(besDeps, resources.New(services.InstanceGroup, &instancegroup.Config{
			PoolName: pool,
			Ports:    cfg.Ports,
		}, nil, func(ensured interface{}) error {
			groupCB(pool, ensured.(*instancegroup.Ensured))
			return nil
		}))
	}

	backendServices := make([]*core.Resource, 0)
	if cfg.External {
		for _, port := range cfg.Ports {
			backendServices = append(backendServices, resources.New(services.BackendService, &backendservice.Config{
				External: &backendservice.External{
					Port: uint16(port),
				},
			}, besDeps, nil))
		}
	} else {
		backendServices = append(backendServices, resources.New(services.BackendService, &backendservice.Config{}, besDeps, nil))
	}

	return backendServices, nil
}
