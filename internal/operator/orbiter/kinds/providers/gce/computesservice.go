package gce

import (
	"fmt"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/providers/core"

	uuid "github.com/satori/go.uuid"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbos/mntr"
	"google.golang.org/api/compute/v1"
)

var _ core.MachinesService = (*machinesService)(nil)

type machinesService struct {
	monitor    mntr.Monitor
	desired    *Spec
	providerID string
	orbID      string
	projectID  string
	client     *compute.Service
	cache      struct {
		instances map[string][]*instance
	}
	onCreate func(pool string, machine infra.Machine)
}

func newMachinesService(
	monitor mntr.Monitor,
	desired *Spec,
	orbID string,
	providerID string,
	projectID string,
	client *compute.Service,
) *machinesService {
	return &machinesService{
		monitor:    monitor,
		desired:    desired,
		providerID: providerID,
		orbID:      orbID,
		projectID:  projectID,
		client:     client,
	}
}

func (m *machinesService) Create(poolName string) (infra.Machine, error) {

	desired, ok := m.desired.Pools[poolName]
	if !ok {
		return nil, fmt.Errorf("Pool %s is not configured", poolName)
	}

	id := newName()

	// Calculate minimum cpu and memory according to the gce specs:
	// https://cloud.google.com/machine/docs/instances/creating-instance-with-custom-machine-type#specifications
	cores := desired.MinCPUCores
	if cores > 1 {
		if cores%2 != 0 {
			cores++
		}
	}
	memory := float64(desired.MinMemoryGB * 1024)
	memoryPerCore := memory / float64(cores)
	minMemPerCore := 922
	maxMemPerCore := 6656
	for memoryPerCore < float64(minMemPerCore) {
		memoryPerCore = memory / float64(cores)
		memory += 256
	}

	for memoryPerCore > float64(maxMemPerCore) {
		cores++
		memoryPerCore = float64(memory) / float64(cores)
	}

	name := newName()
	sshKey := fmt.Sprintf("orbiter:%s", m.desired.SSHKey.Public.Value)
	createInstance := &compute.Instance{
		Name:        name,
		MachineType: fmt.Sprintf("zones/%s/machineTypes/custom-%d-%d", m.desired.Zone, cores, int(memory)),
		NetworkInterfaces: []*compute.NetworkInterface{{
			AccessConfigs: []*compute.AccessConfig{{ // Assigns an ephemeral external ip
				Type: "ONE_TO_ONE_NAT",
			}},
		}},
		Tags: &compute.Tags{
			Items: networkTags(m.orbID, m.providerID, poolName),
		},
		Labels: map[string]string{
			"orb":      m.orbID,
			"provider": m.providerID,
			"pool":     poolName,
		},
		Metadata: &compute.Metadata{
			Items: []*compute.MetadataItems{{
				Key:   "ssh-keys",
				Value: &sshKey,
			}},
		},
		Disks: []*compute.AttachedDisk{{
			AutoDelete: true,
			Boot:       true,
			InitializeParams: &compute.AttachedDiskInitializeParams{
				DiskSizeGb:  int64(desired.StorageGB),
				SourceImage: desired.OSImage,
			}},
		},
		Scheduling: &compute.Scheduling{
			Preemptible: desired.Preemptible,
		},
	}

	monitor := m.monitor.WithFields(map[string]interface{}{
		"name": name,
		"pool": poolName,
		"zone": m.desired.Zone,
	})

	if err := operateFunc(
		func() { monitor.Info("Creating instance") },
		m.client.Instances.Insert(m.projectID, m.desired.Zone, createInstance).RequestId(id).Do,
		nil,
	)(); err != nil {
		return nil, err
	}

	newInstance, err := m.client.Instances.Get(m.projectID, m.desired.Zone, createInstance.Name).
		Fields("selfLink,networkInterfaces(accessConfigs(natIP))").
		Do()
	if err != nil {
		return nil, err
	}

	infraMachine := newMachine(
		m.monitor,
		createInstance.Name,
		newInstance.NetworkInterfaces[0].AccessConfigs[0].NatIP,
		newInstance.SelfLink,
		poolName,
		m.removeMachineFunc(
			poolName,
			createInstance.Name,
		),
	)

	if err := infraMachine.UseKey([]byte(m.desired.SSHKey.Private.Value)); err != nil {
		return nil, err
	}

	if m.cache.instances != nil {
		if _, ok := m.cache.instances[poolName]; !ok {
			m.cache.instances[poolName] = make([]*instance, 0)
		}
		m.cache.instances[poolName] = append(m.cache.instances[poolName], infraMachine)
	}

	m.onCreate(poolName, infraMachine)
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

	instances, err := m.client.Instances.
		List(m.projectID, m.desired.Zone).
		Filter(fmt.Sprintf(`labels.orb=%s AND labels.provider=%s`, m.orbID, m.providerID)).
		Fields("items(name,labels,selfLink,networkInterfaces(accessConfigs(natIP)))").
		Do()
	if err != nil {
		return nil, err
	}

	m.cache.instances = make(map[string][]*instance)
	for _, inst := range instances.Items {
		if inst.Labels["orb"] != m.orbID || inst.Labels["provider"] != m.providerID {
			continue
		}

		pool := inst.Labels["pool"]
		mach := newMachine(
			m.monitor,
			inst.Name,
			inst.NetworkInterfaces[0].AccessConfigs[0].NatIP,
			inst.SelfLink,
			pool,
			m.removeMachineFunc(pool, inst.Name),
		)
		if err := mach.UseKey([]byte(m.desired.SSHKey.Private.Value)); err != nil {
			return nil, err
		}
		m.cache.instances[pool] = append(m.cache.instances[pool], mach)
	}

	return m.cache.instances, nil

}

func (m *machinesService) removeMachineFunc(pool, id string) func() error {
	return func() error {

		cleanMachines := make([]*instance, 0)
		for _, cachedMachine := range m.cache.instances[pool] {
			if cachedMachine.id != id {
				cleanMachines = append(cleanMachines, cachedMachine)
			}
		}
		m.cache.instances[pool] = cleanMachines

		return removeResourceFunc(
			m.monitor.WithField("pool", pool),
			"instance",
			id,
			m.client.Instances.Delete(m.projectID, m.desired.Zone, id).RequestId(uuid.NewV1().String()).Do,
		)()
	}
}

func networkTags(orbID, providerID, poolName string) []string {
	return []string{
		fmt.Sprintf("orb-%s", orbID),
		fmt.Sprintf("provider-%s", providerID),
		fmt.Sprintf("pool-%s", poolName),
	}
}
