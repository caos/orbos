// +build test integration

package instancegroup_test

import (
	"os"
	"testing"

	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/gce/api"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/gce/config"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/gce/resourceservices/instance"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/gce/resourceservices/instancegroup"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/integration/core"
	"github.com/caos/orbiter/logging/base"
	logcontext "github.com/caos/orbiter/logging/context"
)

var configCB func() *core.Vipers

func init() {
	configCB = core.Config()
}

func TestAddMachineWorks(t *testing.T) {
	// TODO: Resolve race conditions
	// t.Parallel()

	cfg := configCB()

	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	testPool := "eliotestpool"
	_, _, assemblyInt, err := core.Gce(cfg.Config.Sub("spec.providers.gce"), cfg.Secrets).Assemble("igtest", []string{testPool}, nil)
	if err != nil {
		panic(err)
	}

	assembly := assemblyInt.(*config.Assembly)

	caller := &api.Caller{
		Ctx: assembly.AppContext(),
		Cfg: assembly.Config(),
	}

	logger := logcontext.Add(stdlib.New(os.Stdout)).Verbose()

	svc := instancegroup.New(assembly.AppContext(), logger, assembly.Config(), caller)
	desired, err := svc.Desire(&instancegroup.Config{
		PoolName: testPool,
		Ports:    []int64{80},
	})
	if err != nil {
		panic(err)
	}

	ensured, err := svc.Ensure("testig", desired, nil)
	if err != nil {
		panic(err)
	}

	iSvc := instance.NewInstanceService(logger, assembly, caller)
	inst, err := iSvc.Create(testPool)
	if err != nil {
		panic(err)
	}

	defer func() {
		if delErr := inst.Remove(); delErr != nil {
			panic(delErr)
		}
	}()

	instances, err := iSvc.List(testPool)
	if err != nil {
		panic(err)
	}
	if len(instances) != 1 {
		panic("Expected exactly one instance")
	}

	if err := ensured.(*instancegroup.Ensured).EnsureMembers(instances); err != nil {
		panic(err)
	}

}
