package gce

import (
	"fmt"
	"sort"
	"strings"

	"github.com/caos/orbiter/mntr"

	uuid "github.com/satori/go.uuid"

	"google.golang.org/api/compute/v1"

	"github.com/caos/orbiter/internal/operator/orbiter/kinds/loadbalancers/dynamic"
)

type normalizedLoadbalancer struct {
	forwardingRule *forwardingRule // unique
	targetPool     *targetPool     // unique
	healthcheck    *healthcheck    // unique
	firewall       *firewall       // unique
	address        *address        // The same externalIP reference appears in multiple normalizedLoadbalancer references
}

type StandardLogFunc func(msg string) func()

type forwardingRule struct {
	log StandardLogFunc
	gce *compute.ForwardingRule
}

type targetPool struct {
	log       func(msg string, instances instances) func()
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

func forwardingRuleDesc(lb *normalizedLoadbalancer) string {
	desc := fmt.Sprintf("orb=%s;provider=%s;transport=%s;port=%d", s.orbID, s.providerID, lb.targetPool.transport, lb.healthcheck.port)
	return desc
}

func addressDesc(lb *normalizedLoadbalancer) string {
	sort.Strings(lb.externalIP.transports)
	desc := fmt.Sprintf("orb=%s;provider=%s;transports=%s", s.orbID, s.providerID, strings.Join(lb.externalIP.transports, "-"))
	return desc
}

// normalize returns a normalizedLoadBalancing for each unique destination port and ip combination
// whereas only the first configured healthcheck is relevant
func normalize(spec map[string][]*dynamic.VIP, orbID, providerID string) []*normalizedLoadbalancer {
	var normalized []*normalizedLoadbalancer

	for _, vips := range spec {
		for _, vip := range vips {
			addVIP := true
			for _, src := range vip.Transport {
				if addVIP {
					for _, normal := range normalized {
						if normal.externalIP.ip == *vip.IP {
							normal.externalIP.name = fmt.Sprintf("%s-%s", normal.externalIP.name, src.Name)
							addVIP = false
						}
					}
				}
				if addVIP {
					normalized = append(normalized, &normalizedLoadbalancer{
						externalIP: &externalIP{
							ip:   *vip.IP,
							name: fmt.Sprintf("%s-%s-%s", orbID, providerID, src.Name),
						},
					})
				}
				for _, dest := range src.Destinations {
					// Targetpool ID
					fmt.Sprintf("%s-%s-%s-%s", orbID, providerID, src.Name, dest.Pool)
				}
			}
		}
	}
}

func newName() string {
	return fmt.Sprintf("orbos-%s", uuid.NewV1().String())
}

func removeLog(monitor mntr.Monitor, resource, id string, removed bool) func() {
	msg := "Removing resource"
	if removed {
		msg = "Resource removed"
	}
	monitor = monitor.WithFields(map[string]interface{}{
		"type": resource,
		"id":   id,
	})
	return func() {
		monitor.Info(msg)
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

type ensureFunc func(*context, []*normalizedLoadbalancer) error

func compose(ensure ensureFunc, next ...ensureFunc) ensureFunc {
	newEnsureFunc := ensure
	for _, fn := range append([]ensureFunc{ensure}, next...) {
		newEnsureFunc = func(ctx *context, lb []*normalizedLoadbalancer) error {
			if err := newEnsureFunc(ctx, lb); err != nil {
				return err
			}
			return fn(ctx, lb)
		}
	}
	return newEnsureFunc
}
