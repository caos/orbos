package ctrlgitops

import (
	"runtime/debug"
	"time"

	"github.com/caos/orbos/internal/operator/boom"
	"github.com/caos/orbos/mntr"
)

func Boom(monitor mntr.Monitor, orbConfigPath string, version string) error {

	ensureClient := gitClient(monitor, "ensure")
	queryClient := gitClient(monitor, "query")

	// We don't need to check both clients
	go checks(monitor, queryClient)

	boom.Metrics(monitor)

	takeoffChan := make(chan struct{})
	go func() {
		takeoffChan <- struct{}{}
	}()

	for range takeoffChan {

		ensureChan := make(chan struct{})
		queryChan := make(chan struct{})

		ensure, query := boom.Takeoff(
			monitor,
			"/boom",
			orbConfigPath,
			ensureClient,
			queryClient,
		)
		go func() {
			started := time.Now()
			query()

			monitor.WithFields(map[string]interface{}{
				"took": time.Since(started),
			}).Info("Iteration done")
			debug.FreeOSMemory()

			queryChan <- struct{}{}
		}()
		go func() {
			started := time.Now()
			ensure()

			monitor.WithFields(map[string]interface{}{
				"took": time.Since(started),
			}).Info("Iteration done")
			debug.FreeOSMemory()

			ensureChan <- struct{}{}
		}()

		go func() {
			<-queryChan
			<-ensureChan

			takeoffChan <- struct{}{}
		}()
	}

	return nil
}
