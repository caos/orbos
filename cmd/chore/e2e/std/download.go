package std

import (
	"context"
	"testing"

	"github.com/afiskon/promtail-client/promtail"

	"github.com/rendon/testcli"

	"github.com/caos/orbos/cmd/chore/orbctl"
)

type programSettings struct {
	ctx                                            context.Context
	logger                                         promtail.Client
	orbID, tag, orbconfig, clusterkey, providerkey string
	from                                           uint8
	cleanup, debugOrbctlCommands, download         bool
	cache                                          struct {
		artifactsVersion string
	}
}

func TestDownload(ctx context.Context, t *testing.T, settings programSettings) {

	cmdFunc, err := orbctl.Command(false, false, true, "cypress-testing-dev")
	if err != nil {
		t.Fatal(err)
	}

	cmd := cmdFunc(context.Background())
	if err := cmd.Run(); err != nil {

	}

	// Using package functions
	testcli.Run("greetings")
	if !testcli.Success() {
		t.Fatalf("Expected to succeed, but failed: %s", testcli.Error())
	}

	if !testcli.StdoutContains("Hello?") {
		t.Fatalf("Expected %q to contain %q", testcli.Stdout(), "Hello?")
	}
}
