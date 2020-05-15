package gce

import (
	"fmt"
	"sort"
	"strings"

	"google.golang.org/api/googleapi"

	"github.com/caos/orbos/internal/operator/orbiter"

	"github.com/caos/orbos/mntr"

	uuid "github.com/satori/go.uuid"

	"google.golang.org/api/compute/v1"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/loadbalancers/dynamic"
)

type normalizedLoadbalancer struct {
	forwardingRule *forwardingRule // unique
	targetPool     *targetPool     // unique
	healthcheck    *healthcheck    // unique
	firewalls      []*firewall     // Each pool has its own firewall
	address        *address        // The same externalIP reference appears in multiple normalizedLoadbalancer references
	transport      string
}

type StandardLogFunc func(msg string, debug bool) func()

type forwardingRule struct {
	log StandardLogFunc
	gce *compute.ForwardingRule
}

type targetPool struct {
	log       func(msg string, debug bool, instances instances) func()
	gce       *compute.TargetPool
	destPools []string
}
type healthcheck struct {
	log StandardLogFunc
	gce *compute.HttpHealthCheck
}
type firewall struct {
	log StandardLogFunc
	gce *compute.Firewall
}
type address struct {
	log StandardLogFunc
	gce *compute.Address
}

type normalizedLoadbalancing []*normalizedLoadbalancer

func (n normalizedLoadbalancing) uniqueAddresses() []*address {
	addresses := make([]*address, 0)
loop:
	for _, lb := range n {
		for _, found := range addresses {
			if lb.address == found {
				continue loop
			}
		}
		addresses = append(addresses, lb.address)
	}
	return addresses
}

// normalize returns a normalizedLoadBalancing for each unique destination port and ip combination
// whereas only one random configured healthcheck is relevant
func normalize(monitor mntr.Monitor, spec map[string][]*dynamic.VIP, orbID, providerID string) []*normalizedLoadbalancer {
	var normalized []*normalizedLoadbalancer

	type normalizedDestination struct {
		port   dynamic.Port
		pools  []string
		hcPath string
	}

	for _, ips := range spec {
		for _, ip := range ips {
			address := &address{}
			addressTransports := make([]string, 0)
			for _, src := range ip.Transport {
				addressTransports = append(addressTransports, src.Name)
				var normalizedDestinations []*normalizedDestination
			normalizeDestinationsLoop:
				for _, dest := range src.Destinations {
					for _, normalizedDest := range normalizedDestinations {
						if dest.Port == normalizedDest.port {
							normalizedDest.pools = append(normalizedDest.pools, dest.Pool)
							continue normalizeDestinationsLoop
						}
					}
					normalizedDestinations = append(normalizedDestinations, &normalizedDestination{
						port:   dest.Port,
						pools:  []string{dest.Pool},
						hcPath: dest.HealthChecks.Path,
					})
				}

				for _, dest := range normalizedDestinations {
					description := fmt.Sprintf("orb=%s;provider=%s;transport=%s;port=%d", orbID, providerID, src.Name, dest.port)
					destMonitor := monitor.WithFields(map[string]interface{}{
						"transport": src.Name,
						"port":      dest.port,
					})
					fwr := &compute.ForwardingRule{
						Description:         description,
						LoadBalancingScheme: "EXTERNAL",
						PortRange:           fmt.Sprintf("%d-%d", dest.port, dest.port),
					}
					tp := &compute.TargetPool{
						Description: description,
					}
					hc := &compute.HttpHealthCheck{
						Description: description,
						Port:        int64(dest.port),
						RequestPath: dest.hcPath,
					}

					poolsLen := len(dest.pools)
					firewalls := make([]*firewall, poolsLen, poolsLen)
					for poolIdx, pool := range dest.pools {
						fw := &compute.Firewall{
							Allowed: []*compute.FirewallAllowed{{
								IPProtocol: "tcp",
								Ports:      []string{fmt.Sprintf("%d", dest.port)},
							}},
							Description:  description,
							SourceRanges: whitelistStrings(src.Whitelist),
							TargetTags:   networkTags(orbID, providerID, pool),
						}
						firewalls[poolIdx] = &firewall{
							log: func(msg string, debug bool) func() {
								localMonitor := destMonitor
								if fw.Name != "" {
									localMonitor = localMonitor.WithField("id", fw.Name)
								}
								level := localMonitor.Info
								if debug {
									level = localMonitor.Debug
								}

								return func() {
									level(msg)
								}
							},
							gce: fw,
						}
					}

					normalized = append(normalized, &normalizedLoadbalancer{
						forwardingRule: &forwardingRule{
							log: func(msg string, debug bool) func() {
								localMonitor := destMonitor
								if fwr.Name != "" {
									localMonitor = localMonitor.WithField("id", fwr.Name)
								}
								level := localMonitor.Info
								if debug {
									level = localMonitor.Debug
								}

								return func() {
									level(msg)
								}
							},
							gce: fwr,
						},
						targetPool: &targetPool{
							log: func(msg string, debug bool, instances instances) func() {
								localMonitor := destMonitor
								if len(instances) > 0 {
									localMonitor = localMonitor.WithField("instances", instances.strings(func(i *instance) string { return i.id }))
								}
								if tp.Name != "" {
									localMonitor = localMonitor.WithField("id", tp.Name)
								}
								level := localMonitor.Info
								if debug {
									level = localMonitor.Debug
								}
								return func() {
									level(msg)
								}
							},
							gce:       tp,
							destPools: dest.pools,
						},
						healthcheck: &healthcheck{
							log: func(msg string, debug bool) func() {
								localMonitor := destMonitor
								if hc.Name != "" {
									localMonitor = localMonitor.WithField("id", hc.Name)
								}
								level := localMonitor.Info
								if debug {
									level = localMonitor.Debug
								}

								return func() {
									level(msg)
								}
							},
							gce: hc,
						},
						firewalls: firewalls,
						address:   address,
						transport: src.Name,
					})
				}
			}
			sort.Strings(addressTransports)
			address.gce = &compute.Address{
				Description: fmt.Sprintf("orb=%s;provider=%s;transports=%s", orbID, providerID, strings.Join(addressTransports, ",")),
			}
			address.log = func(msg string, debug bool) func() {
				localMonitor := monitor.WithField("transports", addressTransports)
				if address.gce.Name != "" {
					localMonitor = localMonitor.WithField("id", address.gce.Name)
				}
				level := localMonitor.Info
				if debug {
					level = localMonitor.Debug
				}

				return func() {
					level(msg)
				}
			}
		}
	}
	return normalized
}

func newName() string {
	return fmt.Sprintf("orbos-%s", uuid.NewV1().String())
}

func removeLog(monitor mntr.Monitor, resource, id string, removed bool, debug bool) func() {
	msg := "Removing resource"
	if removed {
		msg = "Resource removed"
	}
	monitor = monitor.WithFields(map[string]interface{}{
		"type": resource,
		"id":   id,
	})
	level := monitor.Info
	if debug {
		level = monitor.Debug
	}
	return func() {
		level(msg)
	}
}

func removeResourceFunc(monitor mntr.Monitor, resource, id string, call func(...googleapi.CallOption) (*compute.Operation, error)) func() error {
	return func() error {
		if err := operateFunc(
			removeLog(monitor, resource, id, false, true),
			call,
			nil,
		)(); err != nil {
			googleErr, ok := err.(*googleapi.Error)
			if !ok || googleErr.Code != 404 {
				return err
			}
		}
		removeLog(monitor, resource, id, true, false)()
		return nil
	}
}

type context struct {
	monitor         mntr.Monitor
	orbID           string
	providerID      string
	projectID       string
	region          string
	client          *compute.Service
	machinesService *machinesService
}

type queryFunc func(*context, []*normalizedLoadbalancer) ([]func() error, error)

func chain(ctx *context, lb []*normalizedLoadbalancer, query ...queryFunc) (func() error, error) {
	var operations []func() error
	for _, fn := range query {

		ensure, err := fn(ctx, lb)
		if err != nil {
			return nil, err
		}
		operations = append(operations, ensure...)
	}

	return func() error {
		for _, operation := range operations {
			if err := operation(); err != nil {
				return err
			}
		}
		return nil
	}, nil
}

func whitelistStrings(cidrs []*orbiter.CIDR) []string {
	l := len(cidrs)
	wl := make([]string, l, l)
	for idx, cidr := range cidrs {
		wl[idx] = string(*cidr)
	}
	return wl
}
