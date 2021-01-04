package gce

import (
	"fmt"
	"io"
	"strings"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/providers/core"
	"github.com/caos/orbos/pkg/tree"
)

var _ infra.ProviderCurrent = (*Current)(nil)

type Current struct {
	Common  *tree.Common `yaml:",inline"`
	Current struct {
		pools      map[string]infra.Pool `yaml:"-"`
		Ingresses  map[string]*infra.Address
		cleanupped <-chan error `yaml:"-"`
	}
}

func (c *Current) Pools() map[string]infra.Pool {
	return c.Current.pools
}
func (c *Current) Ingresses() map[string]*infra.Address {
	return c.Current.Ingresses
}
func (c *Current) Cleanupped() <-chan error {
	return c.Current.cleanupped
}

func (c *Current) Kubernetes() infra.Kubernetes {
	return infra.Kubernetes{
		Apply: `apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: slow-ext4
provisioner: kubernetes.io/gce-pd
parameters:
  type: pd-standard
  fstype: ext4
  replication-type: none
---
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: fast-ext4
provisioner: kubernetes.io/gce-pd
parameters:
  type: pd-ssd
  fstype: ext4
  replication-type: none
---
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: slow-xfs
provisioner: kubernetes.io/gce-pd
parameters:
  type: pd-standard
  fstype: xfs
  replication-type: none
---
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: fast-xfs
provisioner: kubernetes.io/gce-pd
parameters:
  type: pd-ssd
  fstype: xfs
  replication-type: none
`,
		CloudController: infra.InTreeCloudControllerManager{
			Supported: true,
			CloudConfig: func(machine infra.Machine) io.Reader {
				instance := machine.(*instance)
				ctx := instance.context
				tags := networkTags(ctx.orbID, ctx.providerID, instance.pool)
				return strings.NewReader(fmt.Sprintf(`[Global]
project-id = "%s"
network-name = "%s"
node-instance-prefix = "orbos-"
multizone = false
local-zone = "%s"
container-api-endpoint = "Don't use container API'"
node-tags = "%s"
node-tags = "%s"
node-tags = "%s"
`,
					ctx.projectID,
					ctx.networkName,
					ctx.desired.Zone,
					tags[0],
					tags[1],
					tags[2],
				))
			},
			ProviderName: "gce",
		},
	}
}

func initPools(current *Current, desired *Spec, context *context, normalized []*normalizedLoadbalancer, machines core.MachinesService) error {

	current.Current.pools = make(map[string]infra.Pool)
	for pool := range desired.Pools {
		current.Current.pools[pool] = newInfraPool(pool, context, normalized, machines)
	}

	pools, err := machines.ListPools()
	if err != nil {
		return nil
	}
	for _, pool := range pools {
		// Also return pools that are not configured
		if _, ok := current.Current.pools[pool]; !ok {
			current.Current.pools[pool] = newInfraPool(pool, context, normalized, machines)
		}
	}
	return nil
}
