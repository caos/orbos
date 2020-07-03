package gce

import (
	"fmt"
	"sync"
	"time"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/providers/ssh"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/providers/core"

	uuid "github.com/satori/go.uuid"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	"google.golang.org/api/compute/v1"
)

var _ core.MachinesService = (*machinesService)(nil)

type machinesService struct {
	context           *context
	oneoff            bool
	maintenanceKey    []byte
	maintenanceKeyPub []byte
	cache             struct {
		instances map[string][]*instance
		sync.Mutex
	}
	onCreate func(pool string, machine infra.Machine) error
}

func newMachinesService(context *context, oneoff bool, maintenanceKey []byte, maintenanceKeyPub []byte) *machinesService {
	return &machinesService{
		context:           context,
		oneoff:            oneoff,
		maintenanceKey:    maintenanceKey,
		maintenanceKeyPub: maintenanceKeyPub,
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
	sshKey := fmt.Sprintf("orbiter:%s", m.maintenanceKeyPub)
	createInstance := &compute.Instance{
		Name:        name,
		MachineType: fmt.Sprintf("zones/%s/machineTypes/custom-%d-%d", m.context.desired.Zone, cores, int(memory)),
		Tags:        &compute.Tags{Items: networkTags(m.context.orbID, m.context.providerID, poolName)},
		NetworkInterfaces: []*compute.NetworkInterface{{
			Network: m.context.networkURL,
		}},
		Labels:     map[string]string{"orb": m.context.orbID, "provider": m.context.providerID, "pool": poolName},
		Disks:      disks,
		Scheduling: &compute.Scheduling{Preemptible: desired.Preemptible},
		Metadata: &compute.Metadata{
			Items: []*compute.MetadataItems{{
				Key:   "ssh-keys",
				Value: &sshKey,
			}},
		},
	}

	monitor := m.context.monitor.WithFields(map[string]interface{}{
		"machine": name,
		"pool":    poolName,
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

	var machine machine
	if m.oneoff || m.maintenanceKey == nil || len(m.maintenanceKey) == 0 {
		machine = newGCEMachine(m.context, monitor, createInstance.Name)
	} else {
		sshMachine := ssh.NewMachine(monitor, "orbiter", newInstance.NetworkInterfaces[0].NetworkIP)
		if err := sshMachine.UseKey(m.maintenanceKey); err != nil {
			return nil, err
		}
		machine = sshMachine
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
		if m.oneoff || m.maintenanceKey == nil || len(m.maintenanceKey) == 0 {
			machine = newGCEMachine(m.context, m.context.monitor.WithFields(toFields(inst.Labels)), inst.Name)
		} else {
			sshMachine := ssh.NewMachine(m.context.monitor.WithFields(toFields(inst.Labels)), "orbiter", inst.NetworkInterfaces[0].NetworkIP)
			if err := sshMachine.UseKey(m.maintenanceKey); err != nil {
				return nil, err
			}
			machine = sshMachine
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
