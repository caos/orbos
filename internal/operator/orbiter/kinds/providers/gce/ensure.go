package gce

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/caos/orbos/internal/helpers"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/providers/core"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/loadbalancers/dynamic"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/loadbalancers/dynamic/wrap"

	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	dynamiclbmodel "github.com/caos/orbos/internal/operator/orbiter/kinds/loadbalancers/dynamic"
	"github.com/caos/orbos/internal/push"
	"github.com/pkg/errors"

	"github.com/caos/orbos/internal/operator/orbiter"
	//	externallbmodel "github.com/caos/orbos/internal/operator/orbiter/kinds/loadbalancers/external"
)

func query(
	desired *Spec,
	current *Current,
	lb interface{},
	context *context,
	currentNodeAgents *common.CurrentNodeAgents,
	nodeAgentsDesired *common.DesiredNodeAgents,
	orbiterCommit,
	repoURL,
	repoKey string,
) (ensureFunc orbiter.EnsureFunc, err error) {

	lbCurrent, ok := lb.(*dynamiclbmodel.Current)
	if !ok {
		errors.Errorf("Unknown or unsupported load balancing of type %T", lb)
	}
	normalized, firewalls := normalize(context.monitor, lbCurrent.Current.Spec, context.orbID, context.providerID)

	ensureLB, err := queryResources(context, normalized, firewalls)
	if err != nil {
		return nil, err
	}

	current.Current.Ingresses = make(map[string]*infra.Address)
	for _, lb := range normalized {
		current.Current.Ingresses[lb.transport] = &infra.Address{
			Location:     lb.address.gce.Address,
			BackendPort:  internalPort(lb),
			FrontendPort: externalPort(lb),
		}
	}

	queryNA, installNA := core.NodeAgentFuncs(context.monitor, orbiterCommit, repoURL, repoKey, currentNodeAgents)

	desireNodeAgent := func(pool string, machine infra.Machine) error {

		machineID := machine.ID()
		machineMonitor := context.monitor.WithField("machine", machineID)
		na, _ := nodeAgentsDesired.Get(machineID)
		if na.Software.Health.Config == nil {
			na.Software.Health.Config = make(map[string]string)
		}

		for _, lb := range normalized {
			for _, destPool := range lb.targetPool.destPools {
				if pool == destPool {

					key := fmt.Sprintf(
						"%s:%d%s",
						"0.0.0.0",
						lb.healthcheck.gce.Port,
						lb.healthcheck.gce.RequestPath)

					value := fmt.Sprintf(
						"%d@%s://%s:%d%s",
						lb.healthcheck.desired.Code,
						lb.healthcheck.desired.Protocol,
						machine.IP(),
						internalPort(lb),
						lb.healthcheck.desired.Path)

					if v := na.Software.Health.Config[key]; v != value {
						na.Software.Health.Config[key] = value
						machineMonitor.WithFields(map[string]interface{}{
							"listen": key,
							"checks": value,
						}).Changed("Healthcheck desired")
					}
					fw := common.ToFirewall(map[string]*common.Allowed{
						lb.healthcheck.gce.Description: {
							Port:     fmt.Sprintf("%d", lb.healthcheck.gce.Port),
							Protocol: "tcp",
						},
					})
					if !na.Firewall.Contains(fw) {
						na.Firewall.Merge(fw)
						machineMonitor.WithField("ports", fw.Ports()).Changed("Firewall desired")
					}
				}
			}
		}
		running, err := queryNA(machine)
		if err != nil {
			return err
		}
		if !running {
			return installNA(machine)
		}

		return nil
	}

	context.machinesService.onCreate = desireNodeAgent
	wrappedMachines := wrap.MachinesService(context.machinesService, *lbCurrent, false, nil, func(vip *dynamic.VIP) string {
		for _, transport := range vip.Transport {
			address, ok := current.Current.Ingresses[transport.Name]
			if ok {
				return address.Location
			}
		}
		panic(fmt.Errorf("external address for %v is not ensured", vip))
	})

	return func(psf push.Func) *orbiter.EnsureResult {
		var done bool
		return orbiter.ToEnsureResult(done, helpers.Fanout([]func() error{
			func() error { return ensureIdentityAwareProxyAPIEnabled(context) },
			func() error { return ensureCloudNAT(context) },
			context.machinesService.restartPreemptibleMachines,
			ensureLB,
			func() error {
				pools, err := context.machinesService.ListPools()
				if err != nil {
					return err
				}

				var desireNodeAgents []func() error
				for _, pool := range pools {
					machines, listErr := context.machinesService.List(pool)
					if listErr != nil {
						err = helpers.Concat(err, listErr)
					}
					for _, machine := range machines {
						desireNodeAgents = append(desireNodeAgents, func(p string, m infra.Machine) func() error {
							return func() error {
								return desireNodeAgent(p, m)
							}
						}(pool, machine))
					}
				}
				return helpers.Fanout(desireNodeAgents)()
			},
			func() error {
				var err error
				done, err = wrappedMachines.InitializeDesiredNodeAgents()
				return err
			},
		})())
	}, initPools(current, desired, context, normalized, wrappedMachines)
}

func internalPort(lb *normalizedLoadbalancer) uint16 {
	return lb.backendPort
}

func externalPort(lb *normalizedLoadbalancer) uint16 {
	port, err := strconv.Atoi(strings.Split(lb.forwardingRule.gce.PortRange, "-")[0])
	if err != nil {
		panic(err)
	}
	return uint16(port)
}
