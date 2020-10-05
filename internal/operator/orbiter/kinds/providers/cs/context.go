package cs

import (
	ctxpkg "context"
	"fmt"
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

func buildContext(monitor mntr.Monitor, desired *Spec, orbID, providerID string, oneoff bool) (*context, error) {

	ctx := ctxpkg.Background()

	client := cloudscale.NewClient(&http.Client{
		Timeout: 1 * time.Second,
	})

	client.AuthToken = "jcuatmwfsnfk44mplqqcqbfdcarugbat"

	h := fnv.New32()
	h.Write([]byte(orbID))
	networkName := fmt.Sprintf("orbos-network-%d", h.Sum32())
	newContext := &context{
		monitor:     monitor,
		orbID:       orbID,
		providerID:  providerID,
		desired:     desired,
		client:      client,
		ctx:         ctx,
		networkName: networkName,
	}
	h.Reset()

	newContext.machinesService = newMachinesService(newContext, oneoff)

	return newContext, nil
}
