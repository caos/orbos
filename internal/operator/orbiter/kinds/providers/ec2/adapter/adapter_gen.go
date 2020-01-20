// Code generated by "gen-kindstubs -parentpath=github.com/caos/orbiter/internal/operator/orbiter/kinds/providers -versions=v0 -kind=orbiter.caos.ch/EC2Provider from file gen.go"; DO NOT EDIT.

package adapter

import (
	"context"

	"github.com/caos/orbiter/internal/operator/orbiter"
"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/ec2/model"
)

type Builder interface {
	Build(model.UserSpec, orbiter.NodeAgentUpdater) (model.Config, Adapter, error)
}

type builderFunc func(model.UserSpec, orbiter.NodeAgentUpdater) (model.Config, Adapter, error)

func (b builderFunc) Build(spec model.UserSpec, nodeagent orbiter.NodeAgentUpdater) (model.Config, Adapter, error) {
	return b(spec, nodeagent)
}

type Adapter interface {
	Ensure(context.Context, *orbiter.Secrets, map[string]interface{}) (*model.Current, error)
}

type adapterFunc func(context.Context, *orbiter.Secrets, map[string]interface{}) (*model.Current, error)

func (a adapterFunc) Ensure(ctx context.Context, secrets *orbiter.Secrets, ensuredDependencies map[string]interface{}) (*model.Current, error) {
	return a(ctx, secrets, ensuredDependencies)
}
