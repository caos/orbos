package gce

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"google.golang.org/api/googleapi"

	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/core"

	"github.com/pkg/errors"

	uuid "github.com/satori/go.uuid"
	"google.golang.org/api/option"

	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbiter/mntr"
	"google.golang.org/api/compute/v1"
)

var _ core.MachinesService = (*machinesService)(nil)

type machinesService struct {
	monitor    mntr.Monitor
	desired    *Spec
	providerID string
	orbID      string
	cache      struct {
		client    *compute.Service
		projectID string
		machines  map[string][]*machine
	}
	//	cache      map[string]cachedMachines
}

func NewMachinesService(
	monitor mntr.Monitor,
	desired *Spec,
	providerID string,
	orbID string) *machinesService {
	return &machinesService{
		monitor:    monitor,
		desired:    desired,
		providerID: providerID,
		orbID:      orbID,
	}
}

func (m *machinesService) Create(poolName string) (infra.Machine, error) {

	client, projectID, err := m.context()
	if err != nil {
		return nil, err
	}

	resources, ok := m.desired.Pools[poolName]
	if !ok {
		return nil, fmt.Errorf("Pool %s is not configured", poolName)
	}

	id := uuid.NewV1().String()
	name := fmt.Sprintf("orbos-%s", id)

	// Calculate minimum cpu and memory according to the gce specs:
	// https://cloud.google.com/machine/docs/instances/creating-instance-with-custom-machine-type#specifications
	cores := resources.MinCPUCores
	if cores > 1 {
		if cores%2 != 0 {
			cores++
		}
	}
	memory := float64(resources.MinMemoryGB * 1024)
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

	sshKey := fmt.Sprintf("orbiter:%s", m.desired.SSHKey.Public.Value)
	instance := &compute.Instance{
		Name: name,
		Labels: map[string]string{
			"orb":      m.orbID,
			"provider": m.providerID,
			"pool":     poolName,
		},
		MachineType: fmt.Sprintf("zones/%s/machineTypes/custom-%d-%d", m.desired.Zone, cores, int(memory)),
		NetworkInterfaces: []*compute.NetworkInterface{{
			AccessConfigs: []*compute.AccessConfig{{ // Assigns an ephemeral external ip
				Type: "ONE_TO_ONE_NAT",
			}},
		}},
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
				DiskSizeGb:  int64(resources.StorageGB),
				SourceImage: resources.OSImage,
			}},
		},
	}

	monitor := m.monitor.WithFields(map[string]interface{}{
		"name":    name,
		"pool":    poolName,
		"project": projectID,
		"zone":    m.desired.Zone,
	})

	if err := operate(
		func() { monitor.Info("Creating machine") },
		client.Instances.Insert(projectID, m.desired.Zone, instance).RequestId(id).Do,
	); err != nil {
		return nil, err
	}

	newInstance, err := client.Instances.Get(projectID, m.desired.Zone, instance.Name).
		Fields("networkInterfaces(accessConfigs(natIP))").
		Do()
	if err != nil {
		return nil, err
	}

	infraMachine := newMachine(
		m.monitor,
		instance.Name,
		newInstance.NetworkInterfaces[0].AccessConfigs[0].NatIP,
		poolName,
		m.removeMachineFunc(
			poolName,
			instance.Name,
		),
	)

	if err := infraMachine.UseKey([]byte(m.desired.SSHKey.Private.Value)); err != nil {
		return nil, err
	}

	if m.cache.machines != nil {
		if _, ok := m.cache.machines[poolName]; !ok {
			m.cache.machines[poolName] = make([]*machine, 0)
		}
		m.cache.machines[poolName] = append(m.cache.machines[poolName], infraMachine)
	}

	monitor.Info("Machine created")
	return infraMachine, nil
}

func (m *machinesService) ListPools() ([]string, error) {

	pools, err := m.machines()
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
	pools, err := m.machines()
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

func operate(before func(), call func(...googleapi.CallOption) (*compute.Operation, error)) error {
	before()
	op, err := call()
	if err != nil {
		return err
	}

	if op.Progress < 100 {
		time.Sleep(time.Second)
		return operate(before, call)
	}
	return nil
}

func (m *machinesService) context() (*compute.Service, string, error) {

	client, err := m.client()
	if err != nil {
		return nil, "", err
	}

	projectID, err := m.projectID()
	return client, projectID, err
}

func (m *machinesService) client() (_ *compute.Service, err error) {
	if m.cache.client == nil {
		m.cache.client, err = compute.NewService(context.TODO(), option.WithCredentialsJSON([]byte(m.desired.JSONKey.Value)))
	}
	return m.cache.client, err
}

func (m *machinesService) projectID() (_ string, err error) {
	if m.cache.projectID == "" {
		jsonKey := struct {
			ProjectID string `json:"project_id"`
		}{}
		err = errors.Wrap(json.Unmarshal([]byte(m.desired.JSONKey.Value), &jsonKey), "extracting project id from jsonkey failed")
		m.cache.projectID = jsonKey.ProjectID
	}
	return m.cache.projectID, err
}

func (m *machinesService) machines() (map[string][]*machine, error) {
	if m.cache.machines != nil {
		return m.cache.machines, nil
	}

	client, projectID, err := m.context()
	if err != nil {
		return nil, err
	}

	instances, err := client.Instances.
		List(projectID, m.desired.Zone).
		Filter(fmt.Sprintf("labels.orb:%s AND labels.provider:%s", m.orbID, m.providerID)).
		Fields("items(name,labels,networkInterfaces(accessConfigs(natIP)))").
		Do()
	if err != nil {
		return nil, err
	}

	m.cache.machines = make(map[string][]*machine)
	for _, inst := range instances.Items {
		if inst.Labels["orb"] != m.orbID || inst.Labels["provider"] != m.providerID {
			continue
		}

		pool := inst.Labels["pool"]
		mach := newMachine(
			m.monitor,
			inst.Name,
			inst.NetworkInterfaces[0].AccessConfigs[0].NatIP,
			pool,
			m.removeMachineFunc(pool, inst.Name),
		)
		if err := mach.UseKey([]byte(m.desired.SSHKey.Private.Value)); err != nil {
			return nil, err
		}
		m.cache.machines[pool] = append(m.cache.machines[pool], mach)
	}

	return m.cache.machines, nil

}

func (m *machinesService) removeMachineFunc(pool, id string) func() error {
	return func() error {

		client, projectID, err := m.context()
		if err != nil {
			return err
		}

		if err := operate(
			func() { m.monitor.Info("Deleting machine") },
			client.Instances.Delete(projectID, m.desired.Zone, id).RequestId(uuid.NewV1().String()).Do,
		); err != nil {
			return err
		}
		cleanMachines := make([]*machine, 0)
		for _, cachedMachine := range m.cache.machines[pool] {
			if cachedMachine.id != id {
				cleanMachines = append(cleanMachines, cachedMachine)
			}
		}
		m.cache.machines[pool] = cleanMachines
		return nil

	}
}
