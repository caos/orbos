package instance

import (
	"github.com/caos/orbiter/logging"
	"github.com/caos/orbiter/internal/kinds/clusters/core/infra"
	"github.com/caos/orbiter/internal/kinds/providers/edge/ssh"
	"github.com/caos/orbiter/internal/kinds/providers/gce/edge/api"
	"github.com/caos/orbiter/internal/kinds/providers/gce/model"
	"google.golang.org/api/compute/v1"
)

type Instance interface {
	URL() string
	cacheIPs(internal string, external string)
	infra.Compute
}

type instance struct {
	logger logging.Logger
	infra.Compute
	spec       *model.UserSpec
	caller     *api.Caller
	svc        *compute.InstancesService
	id         string
	internalIP *string
	externalIP *string
	url        string
}

func newInstance(logger logging.Logger, caller *api.Caller, spec *model.UserSpec, svc *compute.InstancesService, id string, url string, remoteUser string) Instance {
	i := &instance{logger.WithFields(map[string]interface{}{
		"type": "instance",
		"name": id,
	}), nil, spec, caller, svc, id, nil, nil, url}
	i.Compute = ssh.NewCompute(logger, i, remoteUser)
	return i
}

func (m *instance) URL() string {
	return m.url
}

func (m *instance) cacheIPs(internal string, external string) {
	m.internalIP = &internal
	m.externalIP = &external
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

func (m *instance) InternalIP() (*string, error) {
	ip, _, err := m.ips()
	return ip, err
}

func (m *instance) ExternalIP() (*string, error) {
	_, ip, err := m.ips()
	return ip, err
}

func (m *instance) ips() (*string, *string, error) {
	if m.externalIP != nil && m.internalIP != nil {
		return m.internalIP, m.externalIP, nil
	}

	interf, err := m.caller.GetResource(m.id, "networkInterfaces(networkIP,accessConfigs(natIP))", []interface{}{
		m.svc.Get(m.spec.Project, m.spec.Zone, m.id),
	})
	if err != nil {
		return nil, nil, err
	}

	instance := interf.(*compute.Instance)

	m.internalIP = &instance.NetworkInterfaces[0].NetworkIP
	m.externalIP = &instance.NetworkInterfaces[0].AccessConfigs[0].NatIP
	return m.internalIP, m.externalIP, nil

}
