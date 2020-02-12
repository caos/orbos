package instance

import (
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/core"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/gce/edge/api"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/gce/model"
	"github.com/caos/orbiter/logging"
	"google.golang.org/api/machine/v1"
)

type Instance interface {
	URL() string
	infra.Machine
}

type instance struct {
	logger logging.Logger
	infra.Machine
	spec   *model.UserSpec
	caller *api.Caller
	svc    *machine.InstancesService
	id     string
	ip     string
	url    string
}

func newInstance(logger logging.Logger, caller *api.Caller, spec *model.UserSpec, svc *machine.InstancesService, id, url, remoteUser, IP string) Instance {
	i := &instance{logger.WithFields(map[string]interface{}{
		"type": "instance",
		"name": id,
	}), nil, spec, caller, svc, id, IP, url}
	i.Machine = ssh.NewMachine(logger, i, remoteUser)
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
