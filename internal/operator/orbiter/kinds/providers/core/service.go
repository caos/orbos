package core

import (
	"fmt"
	"strings"
	"time"

	"github.com/caos/orbiter/mntr"
	"github.com/pkg/errors"
)

type ResourceFactory struct {
	monitor    mntr.Monitor
	operatorID *string
}

func NewResourceFactory(monitor mntr.Monitor, operatorID string) *ResourceFactory {
	return &ResourceFactory{monitor, &operatorID}
}

func (i *ResourceFactory) New(service ResourceService, payload interface{}, dependencies []*Resource, ensured func(interface{}) error) *Resource {
	return newResource(i.operatorID, &loggedResourceService{i.monitor, service, i.operatorID}, payload, dependencies, ensured)
}

func (i *ResourceFactory) IsOperatorManaged(id *string) bool {
	return strings.HasPrefix(*id, *i.operatorID)
}

type loggedResourceService struct {
	monitor    mntr.Monitor
	svc        ResourceService
	operatorID *string
}

func (l *loggedResourceService) Abbreviate() string {
	return l.svc.Abbreviate()
}

func (l *loggedResourceService) Desire(payload interface{}) (interface{}, error) {
	desired, err := l.svc.Desire(payload)
	if err == nil {
		l.monitor.WithFields(map[string]interface{}{
			"payload": payload,
			"desired": desired,
		}).Debug("Resources desired")
	}
	return desired, errors.Wrapf(err, "desiring %s resource for operator %s with payload %#+v failed", l.svc.Abbreviate(), *l.operatorID, payload)
}

func (l *loggedResourceService) Ensure(id string, desired interface{}, ensuredDependencies []interface{}) (ensured interface{}, err error) {
	started := time.Now()
	defer func() {
		if err != nil {
			return
		}
		l.monitor.WithFields(map[string]interface{}{
			"resource": fmt.Sprintf("%#+v", ensured),
			"took":     time.Now().Sub(started),
		}).Debug("Resource ensured")
	}()
	ensured, err = l.svc.Ensure(id, desired, ensuredDependencies)
	return ensured, errors.Wrapf(err, "ensuring resource %s desired %#+v failed", id, desired)
}

func (l *loggedResourceService) AllExisting() ([]string, error) {
	found, err := l.svc.AllExisting()
	if err == nil {
		l.monitor.Verbose().WithFields(map[string]interface{}{
			"result": found,
		}).Debug("Resources queried")
	} else {
		l.monitor.Verbose().WithFields(map[string]interface{}{
			"error": err,
		}).Debug("Querying resources failed")
	}
	return found, errors.Wrapf(err, "querying operator %s's %s resources failed", *l.operatorID, l.svc.Abbreviate())
}

func (l *loggedResourceService) Delete(id string) (err error) {
	started := time.Now()
	defer func() {
		if err != nil {
			return
		}
		l.monitor.WithFields(map[string]interface{}{
			"id":   id,
			"took": time.Now().Sub(started),
		}).Changed("Resource deleted")
	}()
	return errors.Wrapf(l.svc.Delete(id), "deleting resource %s failed", id)
}
