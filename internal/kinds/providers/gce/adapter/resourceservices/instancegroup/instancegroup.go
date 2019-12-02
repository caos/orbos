package instancegroup

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/caos/orbiter/logging"
	"github.com/caos/orbiter/internal/kinds/providers/core"
	"github.com/caos/orbiter/internal/kinds/providers/gce/edge/api"
	"github.com/caos/orbiter/internal/kinds/providers/gce/model"
	"google.golang.org/api/compute/v1"
)

type instanceGroup struct {
	ctx    context.Context
	logger logging.Logger
	spec   *model.UserSpec
	svc    *compute.InstanceGroupsService
	caller *api.Caller
}

func New(ctx context.Context, logger logging.Logger, svc *compute.Service, spec *model.UserSpec, caller *api.Caller) core.ResourceService {
	return &instanceGroup{
		ctx,
		logger.WithFields(map[string]interface{}{"type": "instance group"}),
		spec,
		compute.NewInstanceGroupsService(svc),
		caller,
	}
}

func (i *instanceGroup) Abbreviate() string {
	return "ig"
}

type Config struct {
	PoolName string
	Ports    []int64
}

type Desired struct {
	IG         *compute.InstanceGroup
	NamedPorts []*compute.NamedPort `hash:"set"`
}

func (i *instanceGroup) Desire(config interface{}) (interface{}, error) {
	cfg, ok := config.(*Config)
	if !ok {
		return nil, errors.New("Config must be of type *unmanaged.Config")
	}

	ports := make([]*compute.NamedPort, len(cfg.Ports))
	for idx, port := range cfg.Ports {
		ports[idx] = &compute.NamedPort{
			Name: fmt.Sprintf("port-%s", strconv.FormatInt(port, 10)),
			Port: port,
		}
	}

	return &Desired{
		IG: &compute.InstanceGroup{
			Network:     fmt.Sprintf("projects/%s/global/networks/default", i.spec.Project),
			Description: cfg.PoolName,
		},
		NamedPorts: ports,
	}, nil
}

func (i *instanceGroup) Ensure(id string, desired interface{}, dependencies []interface{}) (interface{}, error) {

	logger := i.logger.WithFields(map[string]interface{}{"name": id})

	if len(dependencies) > 0 {
		return nil, errors.New("Instance groups cannot have dependencies")
	}

	selflink, err := i.caller.GetResourceSelfLink(id, []interface{}{
		i.svc.Get(i.spec.Project, i.spec.Zone, id),
	})
	if err != nil {
		return nil, err
	}

	if selflink != nil {
		return newEnsured(i.ctx, i.logger, i.spec, i.svc, id, *selflink, i.caller), nil
	}

	des := desired.(*Desired)
	ig := *des.IG
	ig.Name = id
	ig.NamedPorts = des.NamedPorts

	op, err := i.caller.RunFirstSuccessful(
		logger.WithFields(map[string]interface{}{
			"name": id,
		}),
		api.Insert,
		i.svc.Insert(i.spec.Project, i.spec.Zone, &ig))

	return newEnsured(i.ctx, i.logger, i.spec, i.svc, id, op.TargetLink, i.caller), err
}

func (i *instanceGroup) Delete(id string) error {
	logger := i.logger.WithFields(map[string]interface{}{"name": id})
	_, err := i.caller.RunFirstSuccessful(logger, api.Delete, i.svc.Delete(i.spec.Project, i.spec.Zone, id))
	return err
}

func (i *instanceGroup) AllExisting() ([]string, error) {
	return i.caller.ListResources(i, []interface{}{
		i.svc.List(i.spec.Project, i.spec.Zone),
	})
}
