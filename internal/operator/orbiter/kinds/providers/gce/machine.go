package gce

import (
	"io"
	"sort"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/providers/core"

	"github.com/caos/orbos/pkg/tree"
	"github.com/pkg/errors"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbos/mntr"
	"google.golang.org/api/compute/v1"
)

var _ infra.Machine = (*instance)(nil)

type machine interface {
	Execute(stdin io.Reader, cmd string) ([]byte, error)
	Shell() error
	WriteFile(path string, data io.Reader, permissions uint16) error
	ReadFile(path string, data io.Writer) error
}

type instance struct {
	mntr.Monitor
	id      string
	ip      string
	url     string
	pool    string
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
}

func newMachine(
	context *context,
	monitor mntr.Monitor,
	id,
	ip,
	url,
	pool string,
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
		id:                   id,
		ip:                   ip,
		url:                  url,
		pool:                 pool,
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

func (c *instance) ID() string {
	return c.id
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

func (c *instance) Remove() error {
	return c.remove()
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
		return nil, errors.Wrap(err, "parsing desired state failed")
	}
	desiredTree.Parsed = desired

	ctx, err := buildContext(monitor, &desired.Spec, orbID, providerID, true)
	if err != nil {
		return nil, err
	}

	return core.ListMachines(ctx.machinesService)
}
