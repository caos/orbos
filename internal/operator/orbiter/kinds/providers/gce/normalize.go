package gce

import (
	"fmt"
	"sort"
	"strings"

	"github.com/caos/orbos/internal/helpers"

	"google.golang.org/api/googleapi"

	"github.com/caos/orbos/internal/operator/orbiter"

	"github.com/caos/orbos/mntr"

	"google.golang.org/api/compute/v1"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/loadbalancers/dynamic"
)

type normalizedLoadbalancer struct {
	forwardingRule *forwardingRule // unique
	targetPool     *targetPool     // unique
	healthcheck    *healthcheck    // unique
	address        *address        // The same externalIP reference appears in multiple normalizedLoadbalancer references
	transport      string
	backendPort    uint16
}

type StandardLogFunc func(msg string, debug bool) func()

type forwardingRule struct {
	log StandardLogFunc
	gce *compute.ForwardingRule
}

type targetPool struct {
	log       func(msg string, debug bool, instances []*instance) func()
	gce       *compute.TargetPool
	destPools []string
}

type healthcheck struct {
	log     StandardLogFunc
	gce     *compute.HttpHealthCheck
	desired dynamic.HealthChecks
	pools   []string
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

func (n normalizedLoadbalancing) Len() int      { return len(n) }
func (n normalizedLoadbalancing) Swap(i, j int) { n[i], n[j] = n[j], n[i] }
func (n normalizedLoadbalancing) Less(i, j int) bool {
	return n[i].forwardingRule.gce.Description < n[j].forwardingRule.gce.Description
}

type normalizedDestination struct {
	port  dynamic.Port
	pools []string
	hc    dynamic.HealthChecks
}

type sortableDestinations []*normalizedDestination

func (n sortableDestinations) Len() int      { return len(n) }
func (n sortableDestinations) Swap(i, j int) { n[i], n[j] = n[j], n[i] }
func (n sortableDestinations) Less(i, j int) bool {
	return n[i].port < n[j].port || n[i].hc.Protocol < n[i].hc.Protocol || n[i].hc.Path < n[i].hc.Path || n[i].hc.Code < n[i].hc.Code
}

// normalize returns a normalizedLoadBalancing for each unique destination backendPort and ip combination
// whereas only one random configured healthcheck is relevant
func normalize(ctx *context, spec map[string][]*dynamic.VIP) ([]*normalizedLoadbalancer, []*firewall) {
	var normalized []*normalizedLoadbalancer
	var firewalls []*firewall

	providerDescription := fmt.Sprintf("orb=%s;provider=%s", ctx.orbID, ctx.providerID)

	for _, ips := range spec {
		for _, ip := range ips {
			address := &address{}
			addressTransports := make([]string, 0)
			for _, src := range ip.Transport {
				addressTransports = append(addressTransports, src.Name)
				destDescription := fmt.Sprintf("%s;transport=%s", providerDescription, src.Name)
				destMonitor := ctx.monitor.WithFields(map[string]interface{}{
					"transport": src.Name,
				})
				fwr := &compute.ForwardingRule{
					Description:         destDescription,
					LoadBalancingScheme: "EXTERNAL",
					PortRange:           fmt.Sprintf("%d-%d", src.FrontendPort, src.FrontendPort),
				}

				tp := &compute.TargetPool{
					Description: destDescription,
				}
				hc := &compute.HttpHealthCheck{
					Description: destDescription,
					RequestPath: src.HealthChecks.Path,
				}

				normalized = append(normalized, &normalizedLoadbalancer{
					backendPort: uint16(src.BackendPort),
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
						log: func(msg string, debug bool, insts []*instance) func() {
							localMonitor := destMonitor
							if len(insts) > 0 {
								localMonitor = localMonitor.WithField("instances", instances(insts).strings(func(i *instance) string { return i.id }))
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
						destPools: src.BackendPools,
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
						gce:     hc,
						desired: src.HealthChecks,
						pools:   src.BackendPools,
					},
					address:   address,
					transport: src.Name,
				})

				firewalls = append(firewalls, toInternalFirewall(&compute.Firewall{
					Network:     ctx.networkURL,
					Description: src.Name,
					Allowed: []*compute.FirewallAllowed{{
						IPProtocol: "tcp",
						Ports:      []string{fmt.Sprintf("%d", src.FrontendPort)},
					}},
					SourceRanges: whitelistStrings(src.Whitelist),
					TargetTags:   networkTags(ctx.orbID, ctx.providerID, src.BackendPools...),
				}, destMonitor))
			}
			sort.Strings(addressTransports)
			address.gce = &compute.Address{
				Description: fmt.Sprintf("orb=%s;provider=%s;transports=%s", ctx.orbID, ctx.providerID, strings.Join(addressTransports, ",")),
			}
			address.log = func(msg string, debug bool) func() {
				localMonitor := ctx.monitor.WithField("transports", addressTransports)
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

	sort.Sort(normalizedLoadbalancing(normalized))

	var hcPort int64 = 6700
	for _, lb := range normalized {
		lb.healthcheck.gce.Port = hcPort
		firewalls = append(firewalls, toInternalFirewall(&compute.Firewall{
			Network: ctx.networkURL,
			Allowed: []*compute.FirewallAllowed{{
				IPProtocol: "tcp",
				Ports:      []string{fmt.Sprintf("%d", hcPort)},
			}},
			SourceRanges: []string{
				// healthcheck sources, see https://cloud.google.com/load-balancing/docs/health-checks#fw-netlb
				"35.191.0.0/16",
				"209.85.152.0/22",
				"209.85.204.0/22",
			},
			TargetTags: networkTags(ctx.orbID, ctx.providerID, lb.healthcheck.pools...),
		}, ctx.monitor))
		hcPort++
	}

	return normalized, append(firewalls, toInternalFirewall(&compute.Firewall{
		Network: ctx.networkURL,
		Allowed: []*compute.FirewallAllowed{{
			IPProtocol: "tcp",
			Ports:      []string{"0-65535"},
		}, {
			IPProtocol: "udp",
			Ports:      []string{"0-65535"},
		}, {
			IPProtocol: "icmp",
		}, {
			IPProtocol: "ipip",
		}},
		Description:  "allow-internal",
		SourceRanges: []string{"10.128.0.0/9"},
		TargetTags:   networkTags(ctx.orbID, ctx.providerID),
	}, ctx.monitor), toInternalFirewall(&compute.Firewall{
		Network: ctx.networkURL,
		Allowed: []*compute.FirewallAllowed{{
			IPProtocol: "tcp",
			Ports:      []string{"22"},
		}},
		Description:  "allow-ssh-through-iap",
		SourceRanges: []string{"35.235.240.0/20"},
		TargetTags:   networkTags(ctx.orbID, ctx.providerID),
	}, ctx.monitor))
}

func toInternalFirewall(fw *compute.Firewall, monitor mntr.Monitor) *firewall {
	return &firewall{
		log: func(msg string, debug bool) func() {
			if fw.Name != "" {
				monitor = monitor.WithField("id", fw.Name)
			}
			level := monitor.Info
			if debug {
				level = monitor.Debug
			}

			return func() {
				level(msg)
			}
		},
		gce: fw,
	}
}

func newName() string {
	return "orbos-" + helpers.RandomStringRunes(6, []rune("abcdefghijklmnopqrstuvwxyz0123456789"))
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
			computeOpCall(call),
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

func queryLB(context *context, normalized []*normalizedLoadbalancer) (func() error, error) {
	lb, err := chainInEnsureOrder(
		context, normalized,
		queryHealthchecks,
		queryTargetPools,
		queryAddresses,
		queryForwardingRules,
	)

	if err != nil {
		return nil, err
	}

	return func() error {
		for _, fn := range lb {
			if err := fn(); err != nil {
				return err
			}
		}
		return nil
	}, nil
}

type ensureLBFunc func(*context, []*normalizedLoadbalancer) ([]func() error, []func() error, error)

type ensureFWFunc func(*context, []*firewall) ([]func() error, []func() error, error)

func chainInEnsureOrder(ctx *context, lb []*normalizedLoadbalancer, query ...ensureLBFunc) ([]func() error, error) {
	var ensureOperations []func() error
	var removeOperations []func() error

	for _, fn := range query {
		ensure, remove, err := fn(ctx, lb)
		if err != nil {
			return nil, err
		}
		ensureOperations = append(ensureOperations, helpers.Fanout(ensure))
		removeOperations = append(removeOperations, helpers.Fanout(remove))
	}

	for i := 0; i < len(removeOperations)/2; i++ {
		j := len(removeOperations) - i - 1
		removeOperations[i], removeOperations[j] = removeOperations[j], removeOperations[i]
	}

	return append(ensureOperations, removeOperations...), nil
}

func whitelistStrings(cidrs []*orbiter.CIDR) []string {
	l := len(cidrs)
	wl := make([]string, l, l)
	for idx, cidr := range cidrs {
		wl[idx] = string(*cidr)
	}
	return wl
}
