package gce

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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
		instances map[string][]*compute.Instance
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

func (m *machinesService) context() (*compute.Service, string, error) {

	client, err := m.client()
	if err != nil {
		return nil, "", err
	}

	projectID, err := m.projectID()
	return client, projectID, err
}

func (m *machinesService) instances() (map[string][]*compute.Instance, error) {
	if m.cache.instances != nil {
		return m.cache.instances, nil
	}

	client, projectID, err := m.context()
	if err != nil {
		return nil, err
	}

	instances, err := client.Instances.
		List(projectID, m.desired.Zone).
		Filter(fmt.Sprintf("labels.orb:%s AND labels.provider:%s", m.orbID, m.providerID)).
		Fields("items(labels)").
		Do()
	if err != nil {
		return nil, err
	}

	m.cache.instances = make(map[string][]*compute.Instance)
	for _, inst := range instances.Items {
		if inst.Labels["orb"] != m.orbID || inst.Labels["provider"] != m.providerID {
			continue
		}

		m.cache.instances[inst.Labels["pool"]] = append(m.cache.instances[inst.Labels["pool"]], inst)
	}

	return m.cache.instances, nil

}

func (m *machinesService) ListPools() ([]string, error) {

	instances, err := m.instances()
	if err != nil {
		return nil, err
	}

	pools := make([]string, 0)
	for pool := range instances {
		pools = append(pools, pool)
	}
	return pools, nil
}

func (m *machinesService) List(poolName string) (infra.Machines, error) {
	return nil, errors.New("Not yet implemented")
	/*
		instances, err := m.instances()
		if err != nil {
			return nil, err
		}


		pool, err := c.cachedPool(poolName)
		if err != nil {
			return nil, err
		}

		return pool.Machines(active), nil*/
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

	instance := &compute.Instance{
		Name: name,
		Labels: map[string]string{
			"orb":      m.orbID,
			"provider": m.providerID,
			"pool":     poolName,
		},
		MachineType:       fmt.Sprintf("zones/%s/machineTypes/custom-%d-%d", m.desired.Zone, cores, int(memory)),
		NetworkInterfaces: []*compute.NetworkInterface{{
			//					AccessConfigs: []*compute.AccessConfig{{ // Assigns an ephemeral external ip
			//						Type: "ONE_TO_ONE_NAT",
			//					}},
		}},
		Disks: []*compute.AttachedDisk{{
			AutoDelete: true,
			Boot:       true,
			InitializeParams: &compute.AttachedDiskInitializeParams{
				DiskSizeGb:  int64(resources.StorageGB),
				SourceImage: resources.OSImage,
			}},
		},
	}

	call := client.Instances.Insert(projectID, m.desired.Zone, instance).RequestId(id)

	monitor := m.monitor.WithFields(map[string]interface{}{
		"name":    name,
		"pool":    poolName,
		"project": projectID,
		"zone":    m.desired.Zone,
	})

	compOp := &compute.Operation{}
	for compOp.Progress < 100 {
		monitor.Info("Creating machine")
		time.Sleep(time.Second)
		compOp, err = call.Do()
		if err != nil {
			return nil, err
		}
	}

	if m.cache.instances != nil {
		if _, ok := m.cache.instances[poolName]; !ok {
			m.cache.instances[poolName] = make([]*compute.Instance, 0)
		}
		m.cache.instances[poolName] = append(m.cache.instances[poolName], instance)
	}

	monitor.Info("Machine created")
	return nil, nil
}
