// +build test integration

package integration_test

import (
	"bufio"
	"bytes"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/caos/orbiter/internal/kinds/clusters/core/infra"
	"github.com/caos/orbiter/internal/kinds/providers/integration/core"
	"github.com/caos/orbiter/internal/core/helpers"
)

func TestComputeExecutions(t *testing.T) {
	// TODO: Resolve race conditions
	// t.Parallel()

	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	operatorID := "itcomp"

	prov := core.ProvidersUnderTest(configCB)

	if err := core.Cleanup(prov, operatorID); err != nil {
		panic(err)
	}

	pools, err := testPools(testPoolArgs(operatorID))
	if err != nil {
		panic(err)
	}

	for _, pool := range pools {
		first, err := pool.AddCompute()
		if err != nil {
			panic(err)
		}

		second, err := pool.AddCompute()
		if err != nil {
			panic(err)
		}

		var wg sync.WaitGroup
		var firstErr error
		var secondErr error
		wg.Add(2)
		go func() {
			firstErr = run(first)
			wg.Done()
		}()

		go func() {
			secondErr = run(second)
			wg.Done()
		}()

		wg.Wait()
		if firstErr != nil {
			panic(firstErr)
		}

		if secondErr != nil {
			panic(secondErr)
		}
	}

	if err := core.Cleanup(prov, operatorID); err != nil {
		panic(err)
	}
}

func run(compute infra.Compute) error {

	msg := "Okay"
	var stdout []byte
	var err error

	if err = helpers.Retry(time.NewTimer(5*time.Minute), 10*time.Second, func() (bool, fmt.Stringer) {
		stdout, err = compute.Execute(nil, nil, fmt.Sprintf("echo \"%s\"", msg))
		return err != nil, helpers.ErrorStringer(err)
	}); err != nil {
		return err
	}

	scanner := bufio.NewScanner(bytes.NewReader(stdout))
	lastLine := ""
	for scanner.Scan() {
		lastLine = scanner.Text()
	}
	if lastLine != msg {
		return fmt.Errorf("Expected stdout: %s\nReceived stdout: \"%s\"", msg, lastLine)
	}
	return nil
}
