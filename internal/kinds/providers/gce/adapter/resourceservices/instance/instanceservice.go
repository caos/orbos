package instance

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/core/helpers"
	"github.com/caos/orbiter/internal/core/operator"
	"github.com/caos/orbiter/internal/core/secret"
	"github.com/caos/orbiter/internal/kinds/clusters/core/infra"
	"github.com/caos/orbiter/internal/kinds/providers/core"
	"github.com/caos/orbiter/internal/kinds/providers/gce/edge/api"
	"github.com/caos/orbiter/internal/kinds/providers/gce/model"
	"github.com/caos/orbiter/logging"
	"google.golang.org/api/compute/v1"
)

type instanceService struct {
	operatorID          string
	spec                *model.UserSpec
	logger              logging.Logger
	ctx                 context.Context
	svc                 *compute.InstancesService
	caller              *api.Caller
	secrets             *operator.Secrets
	newComputePublicKey []byte
	dynamicKeyProperty  string
	fromOutside         bool
}

func NewInstanceService(
	ctx context.Context,
	logger logging.Logger,
	id string,
	svc *compute.Service,
	spec *model.UserSpec,
	caller *api.Caller,
	secrets *operator.Secrets,
	newComputePublicKey []byte,
	dynamicKeyProperty string,
	fromOutside bool) core.ComputesService {
	return &instanceService{
		id,
		spec,
		logger.WithFields(map[string]interface{}{"type": "instance"}),
		ctx, compute.NewInstancesService(svc),
		caller,
		secrets,
		newComputePublicKey,
		dynamicKeyProperty,
		fromOutside}
}

func (i *instanceService) ListPools() ([]string, error) {

	list, err := i.svc.List(
		i.spec.Project,
		i.spec.Zone).
		Filter(fmt.Sprintf("(status=RUNNING) AND (name:%s-*)", i.operatorID)).
		Fields("items(name)").
		Context(i.ctx).
		Do()

	if err != nil {
		return nil, err
	}

	pools := make([]string, 0)
instances:
	for _, instance := range list.Items {
		poolName := strings.Split(instance.Name, "-")[1]
		for _, pool := range pools {
			if poolName == pool {
				continue instances
			}
		}
		pools = append(pools, poolName)
	}

	return pools, nil
}

func (i *instanceService) List(poolName string, active bool) (infra.Computes, error) {
	operator := "="
	if !active {
		operator = "!="
	}
	list, err := i.svc.List(
		i.spec.Project,
		i.spec.Zone).
		Filter(fmt.Sprintf("(status%sRUNNING) AND (name:%s-%s-*)", operator, i.operatorID, poolName)).
		Fields("items(name,selfLink,networkInterfaces(networkIP,accessConfigs(natIP)))").
		Context(i.ctx).
		Do()
	if err != nil {
		return nil, err
	}

	instances := make([]infra.Compute, len(list.Items))
	for idx, inst := range list.Items {
		connect := inst.NetworkInterfaces[0].NetworkIP
		if i.fromOutside {
			connect = inst.NetworkInterfaces[0].AccessConfigs[0].NatIP
		}

		instance := newInstance(i.logger, i.caller, i.spec, i.svc, inst.Name, inst.SelfLink, i.spec.RemoteUser, connect)
		if err := instance.UseKeys(i.secrets, i.dynamicKeyProperty); err != nil && errors.Cause(err) != secret.ErrNotExist {
			return nil, err
		}
		instances[idx] = instance
	}
	return instances, nil
}

func (i *instanceService) Create(poolName string) (infra.Compute, error) {

	resources, ok := i.spec.Pools[poolName]
	if !ok {
		return nil, fmt.Errorf("Pool %s is not configured", poolName)
	}

	if resources.StorageGB < 15 {
		return nil, fmt.Errorf("At least 15 GB disk size is needed, but got %d", resources.StorageGB)
	}

	sshKey := fmt.Sprintf("%s:%s", i.spec.RemoteUser, string(i.newComputePublicKey))

	name := fmt.Sprintf("%s-%s-%s", i.operatorID, poolName, helpers.RandomStringRunes(30, []rune("abcdefghijklmnopqrstuvwxyz0123456789")))[:63]
	logger := i.logger.WithFields(map[string]interface{}{"name": name})

	// Calculate minimum cpu and memory according to the gce specs:
	// https://cloud.google.com/compute/docs/instances/creating-instance-with-custom-machine-type#specifications
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

	op, err := i.caller.RunFirstSuccessful(
		logger,
		api.Insert,
		i.svc.Insert(i.spec.Project, i.spec.Zone, &compute.Instance{
			Name:        name,
			MachineType: fmt.Sprintf("zones/%s/machineTypes/custom-%d-%d", i.spec.Zone, cores, int(memory)),
			NetworkInterfaces: []*compute.NetworkInterface{
				&compute.NetworkInterface{
					AccessConfigs: []*compute.AccessConfig{ // Assigns an ephemeral external ip
						&compute.AccessConfig{
							Type: "ONE_TO_ONE_NAT",
						},
					},
				},
			},

			Metadata: &compute.Metadata{
				Items: []*compute.MetadataItems{
					&compute.MetadataItems{
						Key:   "ssh-keys",
						Value: &sshKey,
					},
				},
			},
			Disks: []*compute.AttachedDisk{
				&compute.AttachedDisk{
					AutoDelete: true,
					Boot:       true,
					InitializeParams: &compute.AttachedDiskInitializeParams{
						DiskSizeGb:  int64(resources.StorageGB),
						SourceImage: resources.OSImage,
					},
				},
			},
		}))
	if err != nil {
		return nil, err
	}

	interf, err := i.caller.GetResource(name, "networkInterfaces(networkIP,accessConfigs(natIP))", []interface{}{
		i.svc.Get(i.spec.Project, i.spec.Zone, name),
	})
	if err != nil {
		return nil, err
	}

	instance := interf.(*compute.Instance)

	connect := instance.NetworkInterfaces[0].NetworkIP
	if i.fromOutside {
		connect = instance.NetworkInterfaces[0].AccessConfigs[0].NatIP
	}

	inst := newInstance(i.logger, i.caller, i.spec, i.svc, name, op.TargetLink, i.spec.RemoteUser, connect)
	return inst, inst.UseKeys(i.secrets, i.dynamicKeyProperty)
}
