package instancegroup

import (
	"context"
	"fmt"

	"github.com/caos/infrop/internal/kinds/clusters/core/infra"
	"github.com/caos/infrop/internal/kinds/providers/core"
	"github.com/caos/infrop/internal/kinds/providers/gce/edge/api"
	"github.com/caos/infrop/internal/kinds/providers/gce/adapter/resourceservices/instance"
	"github.com/caos/infrop/internal/kinds/providers/gce/model"
	"github.com/caos/infrop/internal/core/logging"
	"google.golang.org/api/compute/v1"
)

type Ensured struct {
	ctx    context.Context
	logger logging.Logger
	spec   *model.UserSpec
	svc    *compute.InstanceGroupsService
	name   string
	URL    string
	caller *api.Caller
}

func newEnsured(ctx context.Context, logger logging.Logger, spec *model.UserSpec, svc *compute.InstanceGroupsService, name string, url string, caller *api.Caller) core.EnsuredGroup {
	return &Ensured{ctx, logger.WithFields(map[string]interface{}{
		"type": "instance group",
		"name": name,
	}), spec, svc, name, url, caller}
}

func (e *Ensured) EnsureMembers(computes []infra.Compute) error {

	existing, err := e.svc.ListInstances(
		e.spec.Project,
		e.spec.Zone,
		e.name, &compute.InstanceGroupsListInstancesRequest{InstanceState: "RUNNING"}).
		Context(e.ctx).
		Fields("items(instance)").
		Do()
	if err != nil {
		return err
	}
	e.logger.WithFields(map[string]interface{}{
		"before": len(existing.Items),
		"after":  len(computes),
	}).Debug("Ensuring instances are attached")

	add := make([]*compute.InstanceReference, 0)
	addStr := make([]string, 0)
add:
	for _, comp := range computes {
		in := comp.(instance.Instance)
		for _, item := range existing.Items {
			if item.Instance == in.URL() {
				continue add
			}
		}
		add = append(add, &compute.InstanceReference{Instance: in.URL()})
		addStr = append(addStr, in.ID())
	}

	remove := make([]*compute.InstanceReference, 0)
remove:
	for _, item := range existing.Items {
		for _, comp := range computes {
			out := comp.(instance.Instance)
			if item.Instance == out.URL() {
				continue remove
			}
		}

		remove = append(remove, &compute.InstanceReference{Instance: item.Instance})
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
				&compute.InstanceGroupsAddInstancesRequest{Instances: add})); err != nil {
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
				&compute.InstanceGroupsRemoveInstancesRequest{Instances: remove})); err != nil {
			return err
		}
	}

	return nil
}

func (e *Ensured) AddMember(comp infra.Compute) error {
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
			&compute.InstanceGroupsAddInstancesRequest{Instances: []*compute.InstanceReference{
				&compute.InstanceReference{Instance: instance.URL()},
			}}))
	return err
}
