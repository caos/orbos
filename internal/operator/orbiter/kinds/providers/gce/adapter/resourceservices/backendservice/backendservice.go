package backendservice

import (
	"errors"
	"fmt"

	"github.com/caos/orbiter/logging"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/core"

	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/gce/adapter/resourceservices/healthcheck"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/gce/adapter/resourceservices/instancegroup"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/gce/edge/api"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/gce/model"
	"google.golang.org/api/compute/v1"
)

type backendService struct {
	logger    logging.Logger
	spec      *model.UserSpec
	regionSvc *compute.RegionBackendServicesService
	globalSvc *compute.BackendServicesService
	caller    *api.Caller
}

func New(logger logging.Logger, svc *compute.Service, spec *model.UserSpec, caller *api.Caller) core.ResourceService {
	return &backendService{
		logger:    logger.WithFields(map[string]interface{}{"type": "backend service"}),
		spec:      spec,
		regionSvc: compute.NewRegionBackendServicesService(svc),
		globalSvc: compute.NewBackendServicesService(svc),
		caller:    caller,
	}
}

func (b *backendService) Abbreviate() string {
	return "bes"
}

type External struct {
	Port uint16
}

type Config struct {
	External *External
}

type Desired struct {
	BackendType    *compute.Backend
	BackendService *compute.BackendService
}

func (b *backendService) Desire(config interface{}) (interface{}, error) {

	cfg, ok := config.(*Config)
	if !ok {
		return nil, errors.New("payload must be of type *backendservice.Config")
	}

	var maxConnections int64
	var portName string
	scheme := "INTERNAL"
	if cfg.External != nil {
		scheme = "EXTERNAL"
		maxConnections = 1
		portName = fmt.Sprintf("port-%d", cfg.External.Port)
	}

	return &Desired{
		BackendType: &compute.Backend{
			BalancingMode:  "CONNECTION",
			MaxConnections: maxConnections,
		},
		BackendService: &compute.BackendService{
			Protocol:            "TCP",
			LoadBalancingScheme: scheme,
			PortName:            portName,
		},
	}, nil
}

type Ensured struct {
	URL string
}

func (b *backendService) Ensure(id string, desired interface{}, dependencies []interface{}) (interface{}, error) {

	logger := b.logger.WithFields(map[string]interface{}{"name": id})

	selflink, err := b.caller.GetResourceSelfLink(id, []interface{}{
		b.regionSvc.Get(b.spec.Project, b.spec.Region, id),
		b.globalSvc.Get(b.spec.Project, id),
	})
	if err != nil {
		return nil, err
	}

	if selflink != nil {
		return &Ensured{*selflink}, nil
	}

	cfg := desired.(*Desired)

	backends := make([]*compute.Backend, 0)
	healthchecks := make([]string, 0)
	for _, dep := range dependencies {
		switch typedDep := dep.(type) {
		case *healthcheck.Ensured:
			healthchecks = append(healthchecks, typedDep.URL)
		case *instancegroup.Ensured:
			backend := *cfg.BackendType
			backend.Group = typedDep.URL
			backends = append(backends, &backend)
		default:
			return nil, errors.New("Unknown dependency type")
		}
	}

	bes := *cfg.BackendService
	bes.Name = id
	bes.HealthChecks = healthchecks
	bes.Backends = backends

	var op *compute.Operation
	if bes.LoadBalancingScheme == "INTERNAL" {
		op, err = b.caller.RunFirstSuccessful(
			logger.WithFields(map[string]interface{}{"scope": "regional"}),
			api.Insert, b.regionSvc.Insert(b.spec.Project, b.spec.Region, &bes))
	} else {
		op, err = b.caller.RunFirstSuccessful(
			logger.WithFields(map[string]interface{}{"scope": "global"}),
			api.Insert, b.globalSvc.Insert(b.spec.Project, &bes))
	}
	if err != nil {
		return nil, err
	}
	return &Ensured{op.TargetLink}, nil
}

func (b *backendService) Delete(name string) error {
	logger := b.logger.WithFields(map[string]interface{}{"name": name})
	_, err := b.caller.RunFirstSuccessful(logger, api.Delete,
		b.regionSvc.Delete(b.spec.Project, b.spec.Region, name),
		b.globalSvc.Delete(b.spec.Project, name))
	return err
}

func (b *backendService) AllExisting() ([]string, error) {
	return b.caller.ListResources(b, []interface{}{
		b.regionSvc.List(b.spec.Project, b.spec.Region),
		b.globalSvc.List(b.spec.Project),
	})
}
