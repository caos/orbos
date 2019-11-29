package forwardingrule

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/caos/orbiter/internal/core/logging"
	"github.com/caos/orbiter/internal/kinds/providers/core"
	"github.com/caos/orbiter/internal/kinds/providers/gce/adapter/resourceservices/backendservice"
	"github.com/caos/orbiter/internal/kinds/providers/gce/adapter/resourceservices/targetproxy"
	"github.com/caos/orbiter/internal/kinds/providers/gce/edge/api"
	"github.com/caos/orbiter/internal/kinds/providers/gce/model"
	"google.golang.org/api/compute/v1"
)

type forwardingRule struct {
	logger    logging.Logger
	spec      *model.UserSpec
	regionSvc *compute.ForwardingRulesService
	globalSvc *compute.GlobalForwardingRulesService
	caller    *api.Caller
}

func New(logger logging.Logger, svc *compute.Service, spec *model.UserSpec, caller *api.Caller) core.ResourceService {
	return &forwardingRule{
		logger:    logger.WithFields(map[string]interface{}{"type": "forwarding rule"}),
		spec:      spec,
		regionSvc: compute.NewForwardingRulesService(svc),
		globalSvc: compute.NewGlobalForwardingRulesService(svc),
		caller:    caller,
	}
}

func (f *forwardingRule) Abbreviate() string {
	return "fr"
}

type Config struct {
	External bool
	Ports    []int64
}

type Desired struct {
	Rule  *compute.ForwardingRule
	Ports []string
}

func (f *forwardingRule) Desire(config interface{}) (interface{}, error) {

	cfg, ok := config.(*Config)
	if !ok {
		return nil, errors.New("Payload must be of type *forwardingrule.Config")
	}

	if !cfg.External && (len(cfg.Ports) < 1 || len(cfg.Ports) > 5) {
		return nil, errors.New("If internal, one to five ports must be specified")
	}

	scheme := "INTERNAL"
	if cfg.External {
		scheme = "EXTERNAL"
	}

	ports := make([]string, 0)
	portRanges := make([]string, 0)
	externalAllowed := []string{"25", "43", "110", "143", "195", "443", "465", "587", "700", "993", "995", "1688", "1883", "5222"}
	for _, port := range cfg.Ports {
		portStr := strconv.FormatInt(port, 10)
		if !cfg.External {
			ports = append(ports, portStr)
			continue
		}
		// External
		ok := false
		for _, allowed := range externalAllowed {
			if portStr == allowed {
				portRanges = append(portRanges, fmt.Sprintf("%s-%s", portStr, portStr))
				ok = true
				break
			}
		}
		if !ok {
			return nil, fmt.Errorf("Port must be one of %s", strings.Join(externalAllowed, ","))
		}
	}

	return &compute.ForwardingRule{
		LoadBalancingScheme: scheme,
		Ports:               ports,
		PortRange:           strings.Join(portRanges, ", "),
	}, nil
}

type Ensured struct {
	URL string
	IP  string
}

func (f *forwardingRule) Ensure(id string, desired interface{}, dependencies []interface{}) (interface{}, error) {

	logger := f.logger.WithFields(map[string]interface{}{"name": id})

	existing, err := f.get(id)
	if existing != nil {
		return &Ensured{
			URL: existing.SelfLink,
			IP:  existing.IPAddress,
		}, nil
	}

	if len(dependencies) != 1 {
		return nil, errors.New("Exactly one target dependency must be provided")
	}

	rule := *desired.(*compute.ForwardingRule)
	rule.Name = id

	switch target := dependencies[0].(type) {
	case *backendservice.Ensured:
		if rule.LoadBalancingScheme != "INTERNAL" {
			return nil, errors.New("Scheme must be internal")
		}
		rule.BackendService = target.URL
	case *targetproxy.Ensured:
		if rule.LoadBalancingScheme != "EXTERNAL" {
			return nil, errors.New("Scheme must be external")
		}
		rule.Target = target.URL
	}

	if rule.LoadBalancingScheme == "INTERNAL" {
		_, err = f.caller.RunFirstSuccessful(
			logger.WithFields(map[string]interface{}{
				"scope": "regional",
			}),
			api.Insert,
			f.regionSvc.Insert(f.spec.Project, f.spec.Region, &rule))
	} else {
		_, err = f.caller.RunFirstSuccessful(
			logger.WithFields(map[string]interface{}{
				"scope": "global",
			}),
			api.Insert,
			f.globalSvc.Insert(f.spec.Project, &rule))
	}

	if err != nil {
		return nil, err
	}

	created, err := f.get(id)
	if err != nil {
		return nil, err
	}

	return &Ensured{
		URL: created.SelfLink,
		IP:  created.IPAddress,
	}, nil
}

func (f *forwardingRule) Delete(id string) error {
	logger := f.logger.WithFields(map[string]interface{}{"name": id})
	_, err := f.caller.RunFirstSuccessful(logger, api.Delete,
		f.globalSvc.Delete(f.spec.Project, id),
		f.regionSvc.Delete(f.spec.Project, f.spec.Region, id))
	return err
}

func (f *forwardingRule) AllExisting() ([]string, error) {
	return f.caller.ListResources(f, []interface{}{
		f.globalSvc.List(f.spec.Project),
		f.regionSvc.List(f.spec.Project, f.spec.Region),
	})
}

func (f *forwardingRule) get(id string) (*compute.ForwardingRule, error) {
	found, err := f.caller.GetResource(id, "selfLink,IPAddress", []interface{}{
		f.globalSvc.Get(f.spec.Project, id),
		f.regionSvc.Get(f.spec.Project, f.spec.Region, id),
	})
	if err != nil {
		return nil, err
	}

	if found != nil {
		return found.(*compute.ForwardingRule), nil
	}
	return nil, nil
}
