package cs

import (
	"fmt"
	"sync"

	"github.com/caos/orbos/internal/helpers"

	"github.com/cloudscale-ch/cloudscale-go-sdk"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/providers/ssh"
	"github.com/pkg/errors"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/providers/core"

	uuid "github.com/satori/go.uuid"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
)

var _ core.MachinesService = (*machinesService)(nil)

type machinesService struct {
	context *context
	oneoff  bool
	key     *SSHKey
	cache   struct {
		instances map[string][]*instance
		sync.Mutex
	}
	onCreate func(pool string, machine infra.Machine) error
}

func newMachinesService(context *context, oneoff bool) *machinesService {
	return &machinesService{
		context: context,
		oneoff:  oneoff,
	}
}

func (m *machinesService) use(key *SSHKey) error {
	if key == nil || key.Private == nil || key.Public == nil || key.Private.Value == "" || key.Public.Value == "" {
		return errors.New("machines are not connectable. have you configured the orb by running orbctl configure?")
	}
	m.key = key
	return nil
}

func (m *machinesService) Create(poolName string) (infra.Machine, error) {

	desired, ok := m.context.desired.Pools[poolName]
	if !ok {
		return nil, fmt.Errorf("Pool %s is not configured", poolName)
	}

	name := newName()
	monitor := m.context.monitor.WithFields(map[string]interface{}{
		"machine": name,
		"pool":    poolName,
	})

	monitor.Debug("Creating instance")

	newServer, err := m.context.client.Servers.Create(m.context.ctx, &cloudscale.ServerRequest{
		ZonalResourceRequest:  cloudscale.ZonalResourceRequest{},
		TaggedResourceRequest: cloudscale.TaggedResourceRequest{},
		Name:                  name,
		Flavor:                desired.Flavor,
		Image:                 "centos-7",
		Zone:                  desired.Zone,
		VolumeSizeGB:          30,
		Volumes:               nil,
		Interfaces:            nil,
		BulkVolumeSizeGB:      0,
		SSHKeys:               []string{m.context.desired.SSHKey.Public.Value},
		Password:              "",
		UsePublicNetwork:      boolPtr(m.oneoff),
		UsePrivateNetwork:     boolPtr(true),
		UseIPV6:               boolPtr(false),
		AntiAffinityWith:      "",
		ServerGroups:          nil,
		UserData:              "",
	})
	if err != nil {
		return nil, err
	}

	monitor.Info("Instance created")

	var ip string

	for _, interf := range newServer.Interfaces {
		if m.oneoff && interf.Type == "public" {

			if len(interf.Addresses)
		}
	}
	if m.oneoff {

		machine = newGCEMachine(m.context, monitor, createInstance.Name)
	} else {
		sshMachine :=
		machine = sshMachine
	}

	ssh.NewMachine(monitor, "root", newServer.Interfaces[0].Addresses[0].Address)
	if err := sshMachine.UseKey([]byte(m.key.Private.Value)); err != nil {
		return nil, err
	}

	infraMachine := newMachine(
		m.context,
		monitor,
		createInstance.Name,
		newInstance.NetworkInterfaces[0].NetworkIP,
		newInstance.SelfLink,
		poolName,
		m.removeMachineFunc(
			poolName,
			createInstance.Name,
		),
		false,
		machine,
		false,
		func() {},
		func() {},
		false,
		func() {},
		func() {},
	)

	if m.cache.instances != nil {
		if _, ok := m.cache.instances[poolName]; !ok {
			m.cache.instances[poolName] = make([]*instance, 0)
		}
		m.cache.instances[poolName] = append(m.cache.instances[poolName], infraMachine)
	}

	if err := m.onCreate(poolName, infraMachine); err != nil {
		return nil, err
	}

	monitor.Info("Machine created")
	return infraMachine, nil
}

func (m *machinesService) ListPools() ([]string, error) {

	pools, err := m.instances()
	if err != nil {
		return nil, err
	}

	var poolNames []string
	for poolName := range pools {
		poolNames = append(poolNames, poolName)
	}
	return poolNames, nil
}

func (m *machinesService) List(poolName string) (infra.Machines, error) {
	pools, err := m.instances()
	if err != nil {
		return nil, err
	}

	pool := pools[poolName]
	machines := make([]infra.Machine, len(pool))
	for idx, machine := range pool {
		machines[idx] = machine
	}

	return machines, nil
}

func (m *machinesService) instances() (map[string][]*instance, error) {
	if m.cache.instances != nil {
		return m.cache.instances, nil
	}

	instances, err := m.context.client.Instances.
		List(m.context.projectID, m.context.desired.Zone).
		Filter(fmt.Sprintf(`labels.orb=%s AND labels.provider=%s`, m.context.orbID, m.context.providerID)).
		Fields("items(name,labels,selfLink,status,scheduling(preemptible),networkInterfaces(networkIP))").
		Do()
	if err != nil {
		return nil, err
	}

	m.cache.instances = make(map[string][]*instance)
	for _, inst := range instances.Items {
		if inst.Labels["orb"] != m.context.orbID || inst.Labels["provider"] != m.context.providerID {
			continue
		}

		pool := inst.Labels["pool"]

		var machine machine
		if m.oneoff {
			machine = newGCEMachine(m.context, m.context.monitor.WithFields(toFields(inst.Labels)), inst.Name)
		} else {
			sshMachine := ssh.NewMachine(m.context.monitor.WithFields(toFields(inst.Labels)), "orbiter", inst.NetworkInterfaces[0].NetworkIP)
			if err := sshMachine.UseKey([]byte(m.key.Private.Value)); err != nil {
				return nil, err
			}
			machine = sshMachine
		}

		rebootRequired := false
		unrequireReboot := func() {}
		for idx, req := range m.context.desired.RebootRequired {
			if req == inst.Name {
				rebootRequired = true
				unrequireReboot = func(pos int) func() {
					return func() {
						m.context.desired.RebootRequired = append(m.context.desired.RebootRequired[0:pos], m.context.desired.RebootRequired[pos+1:]...)
					}
				}(idx)
				break
			}
		}

		replacementRequired := false
		unrequireReplacement := func() {}
		for idx, req := range m.context.desired.ReplacementRequired {
			if req == inst.Name {
				replacementRequired = true
				unrequireReplacement = func(pos int) func() {
					return func() {
						m.context.desired.ReplacementRequired = append(m.context.desired.ReplacementRequired[0:pos], m.context.desired.ReplacementRequired[pos+1:]...)
					}
				}(idx)
				break
			}
		}

		mach := newMachine(
			m.context,
			m.context.monitor.WithField("name", inst.Name).WithFields(toFields(inst.Labels)),
			inst.Name,
			inst.NetworkInterfaces[0].NetworkIP,
			inst.SelfLink,
			pool,
			m.removeMachineFunc(pool, inst.Name),
			inst.Status == "TERMINATED" && inst.Scheduling.Preemptible,
			machine,
			rebootRequired,
			func(id string) func() {
				return func() { m.context.desired.RebootRequired = append(m.context.desired.RebootRequired, id) }
			}(inst.Name),
			unrequireReboot,
			replacementRequired,
			func(id string) func() {
				return func() { m.context.desired.ReplacementRequired = append(m.context.desired.ReplacementRequired, id) }
			}(inst.Name),
			unrequireReplacement,
		)

		m.cache.instances[pool] = append(m.cache.instances[pool], mach)
	}

	return m.cache.instances, nil
}

func toFields(labels map[string]string) map[string]interface{} {
	fields := make(map[string]interface{})
	for key, label := range labels {
		fields[key] = label
	}
	return fields
}

func (m *machinesService) removeMachineFunc(pool, id string) func() error {
	return func() error {

		m.cache.Lock()
		cleanMachines := make([]*instance, 0)
		for _, cachedMachine := range m.cache.instances[pool] {
			if cachedMachine.id != id {
				cleanMachines = append(cleanMachines, cachedMachine)
			}
		}
		m.cache.instances[pool] = cleanMachines
		m.cache.Unlock()

		return removeResourceFunc(
			m.context.monitor.WithFields(map[string]interface{}{
				"pool":    pool,
				"machine": id,
			}),
			"instance",
			id,
			m.context.client.Instances.Delete(m.context.projectID, m.context.desired.Zone, id).RequestId(uuid.NewV1().String()).Do,
		)()
	}
}

func networkTags(orbID, providerID string, poolName ...string) []string {
	tags := []string{
		fmt.Sprintf("orb-%s", orbID),
		fmt.Sprintf("provider-%s", providerID),
	}
	for _, pool := range poolName {
		tags = append(tags, fmt.Sprintf("pool-%s", pool))
	}
	return tags
}

func boolPtr(b bool) *bool { return &b }

func newName() string {
	return "orbos-" + helpers.RandomStringRunes(6, []rune("abcdefghijklmnopqrstuvwxyz0123456789"))
}
