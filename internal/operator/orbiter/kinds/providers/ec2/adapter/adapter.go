package adapter

import (
	"context"
	"errors"
	"strings"

	"google.golang.org/api/machine/v1"
	"google.golang.org/api/option"

	"github.com/caos/orbiter/internal/operator/orbiter"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/ec2/model"
	"github.com/caos/orbiter/mntr"
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

func authenticatedService(ctx context.Context, googleApplicationCredentialsValue string) (*machine.Service, error) {
	return machine.NewService(ctx, option.WithCredentialsJSON([]byte(strings.Trim(googleApplicationCredentialsValue, "\""))))
}

func New(monitor mntr.Monitor, id string, lbs map[string]*infra.Ingress, publicKey []byte, privateKeyProperty string) Builder {
	return builderFunc(func(spec model.UserSpec, _ orbiter.NodeAgentUpdater) (model.Config, Adapter, error) {
		return model.Config{}, adapterFunc(func(context.Context, *orbiter.Secrets, map[string]interface{}) (*model.Current, error) {
			return &model.Current{}, errors.New("Not yet implemented")
		}), errors.New("Not yet implemented")
	})
}

func NotifyMaster() string {
	return `#!/bin/bash

set -e

INSTANCE_ID=$1
EIP=$2

aws ec2 disassociate-address --public-ip ${EIP}
aws ec2 associate-address --public-ip ${EIP} --instance-id ${INSTANCE_ID}`
}
