package targetproxy

import (
	"errors"

	"github.com/caos/orbiter/internal/core/logging"
	"github.com/caos/orbiter/internal/kinds/providers/core"
	"github.com/caos/orbiter/internal/kinds/providers/gce/adapter/resourceservices/backendservice"
	"github.com/caos/orbiter/internal/kinds/providers/gce/edge/api"
	"github.com/caos/orbiter/internal/kinds/providers/gce/model"
	"google.golang.org/api/compute/v1"
)

type tp struct {
	logger logging.Logger
	spec   *model.UserSpec
	svc    *compute.TargetTcpProxiesService
	caller *api.Caller
}

func New(logger logging.Logger, svc *compute.Service, spec *model.UserSpec, caller *api.Caller) core.ResourceService {
	return &tp{
		logger: logger.WithFields(map[string]interface{}{"type": "target proxy"}),
		spec:   spec,
		svc:    compute.NewTargetTcpProxiesService(svc),
		caller: caller,
	}
}

func (t *tp) Name() string {
	return "target proxy"
}

func (t *tp) Abbreviate() string {
	return "tp"
}

func (t *tp) Desire(config interface{}) (interface{}, error) {
	if config != nil {
		return nil, errors.New("Target proxies are not configurable")
	}

	return &compute.TargetTcpProxy{
		ProxyHeader: "NONE",
	}, nil
}

type Ensured struct {
	URL string
}

func (t *tp) Ensure(id string, desired interface{}, dependencies []interface{}) (interface{}, error) {

	logger := t.logger.WithFields(map[string]interface{}{"name": id})

	selflink, err := t.caller.GetResourceSelfLink(id, []interface{}{
		t.svc.Get(t.spec.Project, id),
	})
	if err != nil {
		return nil, err
	}

	if selflink != nil {
		return &Ensured{*selflink}, nil
	}

	// ID validations
	if len(dependencies) != 1 {
		return nil, errors.New("target proxies depend on exactly one backend service")
	}

	bes, ok := dependencies[0].(*backendservice.Ensured)
	if !ok {
		return nil, errors.New("target proxies depend on exactly one backend service")
	}

	tp := *desired.(*compute.TargetTcpProxy)
	tp.Name = id
	tp.Service = bes.URL

	op, err := t.caller.RunFirstSuccessful(
		logger,
		api.Insert,
		t.svc.Insert(t.spec.Project, &tp))
	if err != nil {
		return nil, err
	}
	return &Ensured{op.TargetLink}, nil
}

func (t *tp) Delete(id string) error {
	logger := t.logger.WithFields(map[string]interface{}{"name": id})
	_, err := t.caller.RunFirstSuccessful(logger, api.Delete, t.svc.Delete(t.spec.Project, id))
	return err
}

func (t *tp) AllExisting() ([]string, error) {
	return t.caller.ListResources(t, []interface{}{
		t.svc.List(t.spec.Project),
	})
}
