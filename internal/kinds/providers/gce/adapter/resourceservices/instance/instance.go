package instance

import (
	"github.com/caos/orbiter/internal/kinds/clusters/core/infra"
	"github.com/caos/orbiter/internal/kinds/providers/edge/ssh"
	"github.com/caos/orbiter/internal/kinds/providers/gce/edge/api"
	"github.com/caos/orbiter/internal/kinds/providers/gce/model"
	"github.com/caos/orbiter/logging"
	"google.golang.org/api/compute/v1"
)

type Instance interface {
	URL() string
	infra.Compute
}

type instance struct {
	logger logging.Logger
	infra.Compute
	spec   *model.UserSpec
	caller *api.Caller
	svc    *compute.InstancesService
	id     string
	ip     string
	url    string
}

func newInstance(logger logging.Logger, caller *api.Caller, spec *model.UserSpec, svc *compute.InstancesService, id, url, remoteUser, IP string) Instance {
	i := &instance{logger.WithFields(map[string]interface{}{
		"type": "instance",
		"name": id,
	}), nil, spec, caller, svc, id, IP, url}
	i.Compute = ssh.NewCompute(logger, i, remoteUser)
	return i
}

func (m *instance) URL() string {
	return m.url
}

func (m *instance) Remove() error {
	_, err := m.caller.RunFirstSuccessful(
		m.logger,
		api.Delete,
		m.svc.Delete(m.spec.Project, m.spec.Zone, m.id))
	return err
}

func (m *instance) ID() string {
	return m.id
}

func (m *instance) IP() string {
	return m.ip
}
