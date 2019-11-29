// +build test integration

package integration_test

import "github.com/caos/infrop/internal/kinds/providers/integration/core"

var configCB func() *core.Vipers

func init() {
	configCB = core.Config()
}
