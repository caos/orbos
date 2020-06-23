package gce

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbos/mntr"
	"google.golang.org/api/compute/v1"
	"io"
	"sort"
)

var _ infra.Machine = (*instance)(nil)

type machine interface {
	Execute(env map[string]string, stdin io.Reader, cmd string) ([]byte, error)
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
}

func newMachine(context *context, monitor mntr.Monitor, id, ip, url, pool string, remove func() error, start bool, machine machine) *instance {
	return &instance{
		Monitor: monitor,
		id:      id,
		ip:      ip,
		url:     url,
		pool:    pool,
		remove:  remove,
		context: context,
		start:   start,
		machine: machine,
	}
}

func (c *instance) ID() string {
	return c.id
}

func (c *instance) IP() string {
	return c.ip
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
