package adapter

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/caos/infrop/internal/core/logging"
	"github.com/caos/infrop/internal/core/operator"
	"github.com/caos/infrop/internal/kinds/clusters/core/infra"
	"github.com/caos/infrop/internal/kinds/providers/core"
	"github.com/caos/infrop/internal/kinds/providers/static/model"
)

type computesService struct {
	logger               logging.Logger
	spec                 *model.UserSpec
	statusFile           string
	bootstrapKeyProperty string
	dynamicKeyProperty   string
	dynamicPublicKey     []byte
	secrets              *operator.Secrets
	desireHostname       func(compute infra.Compute, pool string) error
}

// TODO: Dont accept the whole spec. Accept exactly the values needed (check other constructors too)
func NewComputesService(
	logger logging.Logger,
	id string,
	spec *model.UserSpec,
	bootstrapKeyProperty string,
	dynamicKeyProperty string,
	dynamicPublicKey []byte,
	secrets *operator.Secrets,
	desireHostname func(compute infra.Compute, pool string) error) core.ComputesService {
	return &computesService{
		logger,
		spec,
		filepath.Join("/var/infrop", id),
		bootstrapKeyProperty,
		dynamicKeyProperty,
		dynamicPublicKey,
		secrets,
		desireHostname,
	}
}

func (c *computesService) ListPools() ([]string, error) {

	pools := make([]string, 0)

	for key := range c.spec.Pools {
		pools = append(pools, key)
	}

	return pools, nil
}

func (c *computesService) List(poolName string, active bool) (infra.Computes, error) {

	cmps, ok := c.spec.Pools[poolName]
	if !ok {
		return nil, fmt.Errorf("Pool %s does not exist", poolName)
	}

	computes := make([]infra.Compute, 0)
	for _, cmp := range cmps {
		var buf bytes.Buffer
		compute := newCompute(c.logger, c.statusFile, c.spec.RemoteUser, &cmp.ID, &cmp.InternalIP, &cmp.ExternalIP)
		if err := compute.UseKeys(c.secrets, c.dynamicKeyProperty, c.bootstrapKeyProperty); err != nil {
			return nil, err
		}
		compute.ReadFile(c.statusFile, &buf)
		isActive := strings.Contains(buf.String(), "active")
		if active && isActive || !active && !isActive {
			computes = append(computes, compute)
		}
	}
	return computes, nil
}

func (c *computesService) Create(poolName string) (infra.Compute, error) {
	cmps, ok := c.spec.Pools[poolName]
	if !ok {
		return nil, fmt.Errorf("Pool %s does not exist", poolName)
	}

	for _, cmp := range cmps {
		var buf bytes.Buffer
		compute := newCompute(c.logger, c.statusFile, c.spec.RemoteUser, &cmp.ID, &cmp.InternalIP, &cmp.ExternalIP)
		if err := compute.UseKeys(c.secrets, c.dynamicKeyProperty, c.bootstrapKeyProperty); err != nil {
			return nil, err
		}
		compute.ReadFile(c.statusFile, &buf)

		if strings.Contains(buf.String(), "active") {
			continue
		}

		if err := compute.WriteFile(c.spec.RemotePublicKeyPath, bytes.NewReader(c.dynamicPublicKey), 600); err != nil {
			return nil, err
		}

		if err := compute.WriteFile(c.statusFile, strings.NewReader("active"), 600); err != nil {
			return nil, err
		}

		return compute, c.desireHostname(compute, poolName)
	}

	return nil, errors.New("No computes left")
}
