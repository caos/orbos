package static

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/caos/orbiter/internal/kinds/clusters/core/infra"
	"github.com/caos/orbiter/internal/kinds/providers/core"
	"github.com/caos/orbiter/logging"
)

type computesService struct {
	logger            logging.Logger
	desired           *DesiredV0
	bootstrapkey      []byte
	maintenancekey    []byte
	maintenancekeyPub []byte
	statusFile        string
	desireHostname    func(compute infra.Compute, pool string) error
}

// TODO: Dont accept the whole spec. Accept exactly the values needed (check other constructors too)
func NewComputesService(
	logger logging.Logger,
	desired *DesiredV0,
	bootstrapkey []byte,
	maintenancekey []byte,
	maintenancekeyPub []byte,
	id string,
	desireHostname func(compute infra.Compute, pool string) error) core.ComputesService {
	return &computesService{
		logger,
		desired,
		bootstrapkey,
		maintenancekey,
		maintenancekeyPub,
		filepath.Join("/var/orbiter", id),
		desireHostname,
	}
}

func (c *computesService) ListPools() ([]string, error) {

	pools := make([]string, 0)

	for key := range c.desired.Spec.Pools {
		pools = append(pools, key)
	}

	return pools, nil
}

func (c *computesService) List(poolName string, active bool) (infra.Computes, error) {

	cmps, ok := c.desired.Spec.Pools[poolName]
	if !ok {
		return nil, fmt.Errorf("Pool %s does not exist", poolName)
	}

	computes := make([]infra.Compute, 0)
	for _, cmp := range cmps {
		var buf bytes.Buffer
		compute := newCompute(c.logger, c.statusFile, c.desired.Spec.RemoteUser, &cmp.ID, cmp.IP)
		if err := compute.UseKey(c.maintenancekey, c.bootstrapkey); err != nil {
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
	cmps, ok := c.desired.Spec.Pools[poolName]
	if !ok {
		return nil, fmt.Errorf("Pool %s does not exist", poolName)
	}

	for _, cmp := range cmps {
		var buf bytes.Buffer
		compute := newCompute(c.logger, c.statusFile, c.desired.Spec.RemoteUser, &cmp.ID, cmp.IP)

		if err := compute.UseKey(c.maintenancekey, c.bootstrapkey); err != nil {
			return nil, err
		}
		compute.ReadFile(c.statusFile, &buf)

		if len(c.maintenancekeyPub) == 0 {
			panic("no maintenancekey")
		}
		if err := compute.WriteFile(c.desired.Spec.RemotePublicKeyPath, bytes.NewReader(c.maintenancekeyPub), 600); err != nil {
			return nil, err
		}

		if strings.Contains(buf.String(), "active") {
			continue
		}

		if err := compute.WriteFile(c.statusFile, strings.NewReader("active"), 600); err != nil {
			return nil, err
		}

		return compute, c.desireHostname(compute, poolName)
	}

	return nil, errors.New("No computes left")
}
