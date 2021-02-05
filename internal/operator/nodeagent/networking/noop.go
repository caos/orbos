package networking

import (
	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/internal/operator/nodeagent"
)

func noopEnsurer() nodeagent.NetworkingEnsurer {
	return nodeagent.NetworkingEnsurerFunc(func(desired common.Networking) (common.NetworkingCurrent, func() error, error) {
		return nil, nil, nil
	})
}
