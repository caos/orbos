// +build test integration

package core

import (
	"context"
	"os"

	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/core"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/static"
	staticconfig "github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/static/config"
	logcontext "github.com/caos/orbiter/logging/context"
	"github.com/caos/orbiter/logging/stdlib"

	"github.com/spf13/viper"
)

type staticProvider struct {
	config  *viper.Viper
	secrets *viper.Viper
}

func Static(config *viper.Viper, secrets *viper.Viper) Provider {
	return &staticProvider{config, secrets}
}

func (s *staticProvider) Assemble(operatorID string, configuredPools []string, configuredLoadBalancers []*LoadBalancer) (infra.Provider, core.MachinesService, interface{}, error) {

	ctx := context.Background()
	assembly, err := staticconfig.New(ctx, s.config, map[string]interface{}{
		"operatorId": operatorID,
	}, s.secrets).Assemble()
	if err != nil {
		return nil, nil, nil, err
	}

	monitor := logcontext.Add(stdlib.New(os.Stdout)).Verbose()
	machinesSvc := static.NewMachinesService(monitor, assembly)

	return static.New(monitor, assembly), machinesSvc, assembly, nil
}
