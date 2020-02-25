package healthcheck

import (
	"errors"

	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/core"
	"github.com/caos/orbiter/mntr"

	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/gce/edge/api"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/gce/model"
	"google.golang.org/api/compute/v1"
)

type hc struct {
	monitor mntr.Monitor
	spec    *model.UserSpec
	svc     *compute.HealthChecksService
	caller  *api.Caller
}

type Config struct {
	Port int64
	Path string
}

func New(monitor mntr.Monitor, svc *compute.Service, spec *model.UserSpec, caller *api.Caller) core.ResourceService {
	return &hc{
		monitor: monitor.WithFields(map[string]interface{}{"type": "health check"}),
		spec:    spec,
		svc:     compute.NewHealthChecksService(svc),
		caller:  caller,
	}
}

func (h *hc) Abbreviate() string {
	return "hc"
}

func (h *hc) Desire(payload interface{}) (interface{}, error) {
	cfg, ok := payload.(*Config)
	if !ok {
		return nil, errors.New("Config must be of type *healthcheck.Config")
	}

	return &machine.HealthCheck{
		Type: "HTTPS",
		HttpsHealthCheck: &machine.HTTPSHealthCheck{
			Port:        cfg.Port,
			RequestPath: cfg.Path,
		},
	}, nil
}

type Ensured struct {
	URL string
}

func (h *hc) Ensure(id string, desired interface{}, dependencies []interface{}) (interface{}, error) {

	monitor := h.monitor.WithFields(map[string]interface{}{"name": id})

	// ID validations
	if len(dependencies) > 0 {
		return nil, errors.New("Healthchecks can't have dependencies")
	}

	selflink, err := h.caller.GetResourceSelfLink(id, []interface{}{
		h.svc.Get(h.spec.Project, id),
	})
	if err != nil {
		return nil, err
	}

	if selflink != nil {
		return &Ensured{*selflink}, nil
	}

	hc := *desired.(*machine.HealthCheck)
	hc.Name = id

	op, err := h.caller.RunFirstSuccessful(
		monitor,
		api.Insert,
		h.svc.Insert(h.spec.Project, &hc))
	if err != nil {
		return nil, err
	}
	return &Ensured{op.TargetLink}, nil
}

func (h *hc) Delete(id string) error {
	monitor := h.monitor.WithFields(map[string]interface{}{"name": id})
	_, err := h.caller.RunFirstSuccessful(monitor, api.Delete, h.svc.Delete(h.spec.Project, id))
	return err
}

func (h *hc) AllExisting() ([]string, error) {
	return h.caller.ListResources(h, []interface{}{
		h.svc.List(h.spec.Project),
	})
}
