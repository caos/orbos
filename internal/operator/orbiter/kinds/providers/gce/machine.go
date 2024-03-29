package gce

import (
	"fmt"
	"io"
	"sort"

	"google.golang.org/api/compute/v1"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/loadbalancers"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/providers/core"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/tree"
)

var _ infra.Machine = (*instance)(nil)

type machine interface {
	Execute(stdin io.Reader, cmd string) ([]byte, error)
	Shell() error
	WriteFile(path string, data io.Reader, permissions uint16) error
	ReadFile(path string, data io.Writer) error
	Zone() string
}

type instance struct {
	mntr.Monitor
	ip      string
	url     string
	pool    string
	zone    string
	remove  func() error
	context *context
	start   bool
	machine
	rebootRequired       bool
	requireReboot        func()
	unrequireReboot      func()
	replacementRequired  bool
	requireReplacement   func()
	unrequireReplacement func()
	X_ID                 string `header:"id"`
	X_internalIP         string `header:"internal ip"`
	X_externalIP         string `header:"external ip"`
	X_Pool               string `header:"pool"`
}

func newMachine(
	context *context,
	monitor mntr.Monitor,
	id,
	ip,
	url,
	pool string,
	zone string,
	remove func() error,
	start bool,
	machine machine,
	rebootRequired bool,
	requireReboot func(),
	unrequireReboot func(),
	replacementRequired bool,
	requireReplacement func(),
	unrequireReplacement func()) *instance {
	return &instance{
		Monitor:              monitor,
		X_ID:                 id,
		ip:                   ip,
		url:                  url,
		pool:                 pool,
		zone:                 zone,
		remove:               remove,
		context:              context,
		start:                start,
		machine:              machine,
		rebootRequired:       rebootRequired,
		requireReboot:        requireReboot,
		unrequireReboot:      unrequireReboot,
		replacementRequired:  replacementRequired,
		requireReplacement:   requireReplacement,
		unrequireReplacement: unrequireReplacement,
	}
}

func (c *instance) Zone() string {
	return c.zone
}

func (c *instance) ID() string {
	return c.X_ID
}

func (c *instance) IP() string {
	return c.ip
}

func (c *instance) RebootRequired() (bool, func(), func()) {
	return c.rebootRequired, c.requireReboot, c.unrequireReboot
}

func (c *instance) ReplacementRequired() (bool, func(), func()) {
	return c.replacementRequired, c.requireReplacement, c.unrequireReplacement
}

func (c *instance) Destroy() (func() error, error) {
	return c.remove, nil
}

type instances []*instance

func (c instances) Len() int           { return len(c) }
func (c instances) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c instances) Less(i, j int) bool { return c[i].ID() < c[j].ID() }

func (i instances) strings(field func(i *instance) string) []string {
	sort.Sort(i)
	l := len(i)
	ret := make([]string, l, l)
	for idx, i := range i {
		ret[idx] = field(i)
	}
	return ret
}

func (i instances) refs() []*compute.InstanceReference {
	sort.Sort(i)
	l := len(i)
	ret := make([]*compute.InstanceReference, l, l)
	for idx, i := range i {
		ret[idx] = &compute.InstanceReference{Instance: i.url}
	}
	return ret
}

func ListMachines(monitor mntr.Monitor, desiredTree *tree.Tree, orbID, providerID string) (map[string]infra.Machine, error) {
	desired, err := parseDesiredV0(desiredTree)
	if err != nil {
		return nil, fmt.Errorf("parsing desired state failed: %w", err)
	}
	desiredTree.Parsed = desired

	_, _, _, _, _, err = loadbalancers.GetQueryAndDestroyFunc(monitor, nil, desired.Loadbalancing, &tree.Tree{}, nil)
	if err != nil {
		return nil, err
	}

	svc, err := service(monitor, &desired.Spec, orbID, providerID, true)
	if err != nil {
		return nil, err
	}

	return core.ListMachines(svc)
}
