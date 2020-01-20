// Code generated by "gen-kindstubs -parentpath=github.com/caos/orbiter/internal/operator/orbiter/kinds/loadbalancers -versions=v0 -kind=orbiter.caos.ch/ExternalLoadBalancer from file gen.go"; DO NOT EDIT.

package v0

import (
	"errors"

	"github.com/caos/orbiter/internal/operator/orbiter"
"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/loadbalancers/external/model"
)

var build func(map[string]interface{}, *orbiter.Secrets, interface{}) (model.UserSpec, func(model.Config) ([]orbiter.Assembler, error))

func Build(spec map[string]interface{}, secrets *orbiter.Secrets, dependant interface{}) (model.UserSpec, func(cfg model.Config) ([]orbiter.Assembler, error)) {
	if build != nil {
		return build(spec, secrets, dependant)
	}
	return model.UserSpec{}, func(_ model.Config) ([]orbiter.Assembler, error) {
		return nil, errors.New("Version v0 for kind orbiter.caos.ch/ExternalLoadBalancer is not yet supported")
	}
}
