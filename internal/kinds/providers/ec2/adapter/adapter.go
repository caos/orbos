package adapter

import (
	"context"
	"errors"
	"strings"

	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"

	"github.com/caos/orbiter/internal/core/operator"
	"github.com/caos/orbiter/internal/kinds/clusters/core/infra"
	"github.com/caos/orbiter/internal/kinds/providers/ec2/model"
	"github.com/caos/orbiter/logging"
)

type infraCurrent struct {
	pools map[string]infra.Pool
	ing   map[string]infra.Address
	cu    <-chan error
}

func (i *infraCurrent) Pools() map[string]infra.Pool {
	return i.pools
}

func (i *infraCurrent) Ingresses() map[string]infra.Address {
	return i.ing
}

func (i *infraCurrent) Cleanupped() <-chan error {
	return i.cu
}

func authenticatedService(ctx context.Context, googleApplicationCredentialsValue string) (*compute.Service, error) {
	return compute.NewService(ctx, option.WithCredentialsJSON([]byte(strings.Trim(googleApplicationCredentialsValue, "\""))))
}

func New(logger logging.Logger, id string, lbs map[string]*infra.Ingress, publicKey []byte, privateKeyProperty string) Builder {
	return builderFunc(func(spec model.UserSpec, _ operator.NodeAgentUpdater) (model.Config, Adapter, error) {
		return model.Config{}, adapterFunc(func(context.Context, *operator.Secrets, map[string]interface{}) (*model.Current, error) {
			return &model.Current{}, errors.New("Not yet implemented")
		}), errors.New("Not yet implemented")
	})
}

func NotifyMaster() string {
	return `#!/bin/bash
aws ec2 disassociate-address --public-ip {{ $vip.IP }}
aws ec2 associate-address --public-ip {{ $vip.IP }} --instance-id {{ $root.Self.ID }}`
}
