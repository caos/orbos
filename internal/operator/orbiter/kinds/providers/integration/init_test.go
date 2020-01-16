// +build test integration

package integration_test

import "github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/integration/core"

var configCB func() *core.Vipers

func init() {
	configCB = core.Config()
}
