package cs

import (
	ctxpkg "context"
	"hash/fnv"
	"net/http"
	"time"

	"github.com/cloudscale-ch/cloudscale-go-sdk"

	"github.com/caos/orbos/mntr"
)

type context struct {
	monitor         mntr.Monitor
	networkName     string
	orbID           string
	providerID      string
	desired         *Spec
	client          *cloudscale.Client
	machinesService *machinesService
	ctx             ctxpkg.Context
}

func buildContext(monitor mntr.Monitor, desired *Spec, orbID, providerID string, oneoff bool) *context {

	ctx := ctxpkg.Background()

	client := cloudscale.NewClient(&http.Client{
		Timeout: 30 * time.Second,
	})

	client.AuthToken = desired.APIToken.Value

	h := fnv.New32()
	h.Write([]byte(orbID))
	newContext := &context{
		monitor:    monitor,
		orbID:      orbID,
		providerID: providerID,
		desired:    desired,
		client:     client,
		ctx:        ctx,
	}
	h.Reset()

	newContext.machinesService = newMachinesService(newContext, oneoff)

	return newContext
}
