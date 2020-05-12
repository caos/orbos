package gce

import (
	"sort"

	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/ssh"
	"github.com/caos/orbiter/mntr"
	"google.golang.org/api/compute/v1"
)

var _ infra.Machine = (*instance)(nil)

type instance struct {
	mntr.Monitor
	id     string
	ip     string
	url    string
	pool   string
	remove func() error
	*ssh.Machine
}

func newMachine(monitor mntr.Monitor, id, ip, url, pool string, remove func() error) *instance {
	return &instance{
		Monitor: monitor,
		id:      id,
		ip:      ip,
		url:     url,
		pool:    pool,
		remove:  remove,
		Machine: ssh.NewMachine(monitor, "orbiter", ip),
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
