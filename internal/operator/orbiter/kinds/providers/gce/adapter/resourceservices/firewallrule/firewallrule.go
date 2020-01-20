package firewallrule

import (
	"errors"

	"github.com/caos/orbiter/logging"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/core"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/gce/edge/api"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/gce/model"
	"google.golang.org/api/compute/v1"
)

type fw struct {
	logger logging.Logger
	spec   *model.UserSpec
	svc    *compute.FirewallsService
	caller *api.Caller
}
type Config struct {
	AllowedPorts []string
	DeniedPorts  []string
	IPRanges     []string
	Egress       bool
}

func New(logger logging.Logger, svc *compute.Service, spec *model.UserSpec, caller *api.Caller) core.ResourceService {
	return &fw{
		spec:   spec,
		logger: logger.WithFields(map[string]interface{}{"type": "firewall rule"}),
		svc:    compute.NewFirewallsService(svc),
		caller: caller,
	}
}

func (f *fw) Abbreviate() string {
	return "fwr"
}

func (f *fw) Desire(payload interface{}) (interface{}, error) {
	cfg, ok := payload.(*Config)
	if !ok {
		return nil, errors.New("Config must be of type *firewallrule.Config")
	}

	if len(cfg.AllowedPorts) > 0 && len(cfg.DeniedPorts) > 0 {
		return nil, errors.New("Cannot specify both allowed and denied in the same time")
	}

	direction := "INGRESS"
	if cfg.Egress {
		direction = "EGRESS"
	}

	var allowed []*compute.FirewallAllowed
	var denied []*compute.FirewallDenied

	if len(cfg.AllowedPorts) > 0 {
		allowed = []*compute.FirewallAllowed{
			&compute.FirewallAllowed{
				IPProtocol: "tcp",
				Ports:      cfg.AllowedPorts,
			},
		}
	}

	if len(cfg.DeniedPorts) > 0 {
		denied = []*compute.FirewallDenied{
			&compute.FirewallDenied{
				IPProtocol: "tcp",
				Ports:      cfg.DeniedPorts,
			},
		}
	}

	return &compute.Firewall{
		Allowed:      allowed,
		Denied:       denied,
		Direction:    direction,
		SourceRanges: cfg.IPRanges,
	}, nil
}

type Ensured struct {
	URL string
}

func (f *fw) Ensure(id string, desired interface{}, dependencies []interface{}) (interface{}, error) {

	logger := f.logger.WithFields(map[string]interface{}{"name": id})

	selflink, err := f.caller.GetResourceSelfLink(id, []interface{}{
		f.svc.Get(f.spec.Project, id),
	})
	if err != nil {
		return nil, err
	}

	if selflink != nil {
		return &Ensured{*selflink}, nil
	}

	fwr := *desired.(*compute.Firewall)
	fwr.Name = id

	op, err := f.caller.RunFirstSuccessful(logger, api.Insert, f.svc.Insert(f.spec.Project, &fwr))
	if err != nil {
		return nil, err
	}
	return &Ensured{op.TargetLink}, nil
}

func (f *fw) Delete(id string) error {
	logger := f.logger.WithFields(map[string]interface{}{"name": id})
	_, err := f.caller.RunFirstSuccessful(logger, api.Delete, f.svc.Delete(f.spec.Project, id))
	return err
}

func (f *fw) AllExisting() ([]string, error) {
	return f.caller.ListResources(f, []interface{}{
		f.svc.List(f.spec.Project),
	})
}
