package gce

import (
	"fmt"
	"time"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/providers/core"

	uuid "github.com/satori/go.uuid"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	"google.golang.org/api/compute/v1"
)

var _ core.MachinesService = (*machinesService)(nil)

type machinesService struct {
	context *context
	cache   struct {
		instances map[string][]*instance
	}
	onCreate func(pool string, machine infra.Machine) error
}

func newMachinesService(context *context) *machinesService {
	return &machinesService{
		context: context,
	}
}

func (m *machinesService) restartPreemptibleMachines() error {
	pools, err := m.instances()
	if err != nil {
		return err
	}

	for _, pool := range pools {
		for _, instance := range pool {
			if instance.start {
				if err := operateFunc(
					func() { instance.Monitor.Debug("Restarting preemptible instance") },
					computeOpCall(m.context.client.Instances.Start(m.context.projectID, m.context.desired.Zone, instance.ID()).RequestId(uuid.NewV1().String()).Do),
					func() error { instance.Monitor.Info("Preemptible instance restarted"); return nil },
				)(); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (m *machinesService) Create(poolName string) (infra.Machine, error) {

	desired, ok := m.context.desired.Pools[poolName]
	if !ok {
		return nil, fmt.Errorf("Pool %s is not configured", poolName)
	}

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

	disks := []*compute.AttachedDisk{{
		Type:       "PERSISTENT",
		AutoDelete: true,
		Boot:       true,
		InitializeParams: &compute.AttachedDiskInitializeParams{
			DiskSizeGb:  int64(desired.StorageGB),
			SourceImage: desired.OSImage,
		}},
	}

	diskNames := make([]string, desired.LocalSSDs)
	for i := 0; i < int(desired.LocalSSDs); i++ {
		name := fmt.Sprintf("nvme0n%d", i+1)
		disks = append(disks, &compute.AttachedDisk{
			Type:       "SCRATCH",
			AutoDelete: true,
			Boot:       false,
			Interface:  "NVME",
			InitializeParams: &compute.AttachedDiskInitializeParams{
				DiskType: fmt.Sprintf("zones/%s/diskTypes/local-ssd", m.context.desired.Zone),
			},
			DeviceName: name,
		})
		diskNames[i] = name
	}

	name := newName()
	createInstance := &compute.Instance{
		Name:              name,
		MachineType:       fmt.Sprintf("zones/%s/machineTypes/custom-%d-%d", m.context.desired.Zone, cores, int(memory)),
		Tags:              &compute.Tags{Items: networkTags(m.context.orbID, m.context.providerID, poolName)},
		NetworkInterfaces: []*compute.NetworkInterface{{}},
		Labels:            map[string]string{"orb": m.context.orbID, "provider": m.context.providerID, "pool": poolName},
		Disks:             disks,
		Scheduling:        &compute.Scheduling{Preemptible: desired.Preemptible},
	}

	monitor := m.context.monitor.WithFields(map[string]interface{}{
		"name": name,
		"pool": poolName,
		"zone": m.context.desired.Zone,
	})

	if err := operateFunc(
		func() { monitor.Debug("Creating instance") },
		computeOpCall(m.context.client.Instances.Insert(m.context.projectID, m.context.desired.Zone, createInstance).RequestId(uuid.NewV1().String()).Do),
		func() error { monitor.Info("Instance created"); return nil },
	)(); err != nil {
		return nil, err
	}

	newInstance, err := m.context.client.Instances.Get(m.context.projectID, m.context.desired.Zone, createInstance.Name).
		Fields("selfLink,networkInterfaces(networkIP)").
		Do()
	if err != nil {
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
	)

	for _, name := range diskNames {
		mountPoint := fmt.Sprintf("/mnt/disks/%s", name)
		if err := infra.Try(monitor, time.NewTimer(time.Minute), 10*time.Second, infraMachine, func(m infra.Machine) error {
			_, formatErr := m.Execute(
				nil,
				nil,
				fmt.Sprintf("sudo mkfs.ext4 -F /dev/%s && sudo mkdir -p /mnt/disks/%s && sudo mount /dev/%s %s && sudo chmod a+w %s && echo UUID=`sudo blkid -s UUID -o value /dev/disk/by-id/google-%s` %s ext4 discard,defaults,nofail 0 2 | sudo tee -a /etc/fstab", name, name, name, mountPoint, mountPoint, name, mountPoint),
			)
			return formatErr
		}); err != nil {
			if cleanupErr := infraMachine.Remove(); cleanupErr != nil {
				panic(cleanupErr)
			}
			return nil, err
		}
		monitor.WithField("mountpoint", mountPoint).Info("Disk formatted")
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
		mach := newMachine(
			m.context,
			m.context.monitor.WithField("name", inst.Name).WithFields(toFields(inst.Labels)),
			inst.Name,
			inst.NetworkInterfaces[0].NetworkIP,
			inst.SelfLink,
			pool,
			m.removeMachineFunc(pool, inst.Name),
			inst.Status == "TERMINATED" && inst.Scheduling.Preemptible,
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

		cleanMachines := make([]*instance, 0)
		for _, cachedMachine := range m.cache.instances[pool] {
			if cachedMachine.id != id {
				cleanMachines = append(cleanMachines, cachedMachine)
			}
		}
		m.cache.instances[pool] = cleanMachines

		return removeResourceFunc(
			m.context.monitor.WithField("pool", pool),
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
