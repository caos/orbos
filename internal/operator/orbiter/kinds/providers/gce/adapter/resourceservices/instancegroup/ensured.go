package instancegroup

import (
	"context"
	"fmt"

	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/core"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/gce/adapter/resourceservices/instance"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/gce/edge/api"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/gce/model"
	"github.com/caos/orbiter/logging"
	"google.golang.org/api/machine/v1"
)

type Ensured struct {
	ctx    context.Context
	logger logging.Logger
	spec   *model.UserSpec
	svc    *machine.InstanceGroupsService
	name   string
	URL    string
	caller *api.Caller
}

func newEnsured(ctx context.Context, logger logging.Logger, spec *model.UserSpec, svc *machine.InstanceGroupsService, name string, url string, caller *api.Caller) core.EnsuredGroup {
	return &Ensured{ctx, logger.WithFields(map[string]interface{}{
		"type": "instance group",
		"name": name,
	}), spec, svc, name, url, caller}
}

func (e *Ensured) EnsureMembers(machines []infra.Machine) error {

	existing, err := e.svc.ListInstances(
		e.spec.Project,
		e.spec.Zone,
		e.name, &machine.InstanceGroupsListInstancesRequest{InstanceState: "RUNNING"}).
		Context(e.ctx).
		Fields("items(instance)").
		Do()
	if err != nil {
		return err
	}
	e.logger.WithFields(map[string]interface{}{
		"before": len(existing.Items),
		"after":  len(machines),
	}).Debug("Ensuring instances are attached")

	add := make([]*machine.InstanceReference, 0)
	addStr := make([]string, 0)
add:
	for _, comp := range machines {
		in := comp.(instance.Instance)
		for _, item := range existing.Items {
			if item.Instance == in.URL() {
				continue add
			}
		}
		add = append(add, &machine.InstanceReference{Instance: in.URL()})
		addStr = append(addStr, in.ID())
	}

	remove := make([]*machine.InstanceReference, 0)
remove:
	for _, item := range existing.Items {
		for _, comp := range machines {
			out := comp.(instance.Instance)
			if item.Instance == out.URL() {
				continue remove
			}
		}

		remove = append(remove, &machine.InstanceReference{Instance: item.Instance})
	}

	if len(add) > 0 {
		if _, err = e.caller.RunFirstSuccessful(
			e.logger.WithFields(map[string]interface{}{
				"instances": fmt.Sprintf("%v", add),
			}),
			api.Add,
			e.svc.AddInstances(
				e.spec.Project,
				e.spec.Zone,
				e.name,
				&machine.InstanceGroupsAddInstancesRequest{Instances: add})); err != nil {
			return err
		}

	}

	if len(remove) > 0 {
		if _, err = e.caller.RunFirstSuccessful(
			e.logger.WithFields(map[string]interface{}{
				"instances": fmt.Sprintf("%v", remove),
			}),
			api.Remove,
			e.svc.RemoveInstances(
				e.spec.Project,
				e.spec.Zone,
				e.name,
				&machine.InstanceGroupsRemoveInstancesRequest{Instances: remove})); err != nil {
			return err
		}
	}

	return nil
}

func (e *Ensured) AddMember(comp infra.Machine) error {
	instance := comp.(instance.Instance)
	_, err := e.caller.RunFirstSuccessful(
		e.logger.WithFields(map[string]interface{}{
			"instances": fmt.Sprintf("%v", []string{instance.ID()}),
		}),
		api.Add,
		e.svc.AddInstances(
			e.spec.Project,
			e.spec.Zone,
			e.name,
			&machine.InstanceGroupsAddInstancesRequest{Instances: []*machine.InstanceReference{
				&machine.InstanceReference{Instance: instance.URL()},
			}}))
	return err
}
