package orb

import (
	"github.com/caos/orbiter/internal/core/operator/orbiter"
)

func AdaptFunc() orbiter.AdaptFunc {
	return func(rawDesired []byte, rawSecrets []byte, nodeAgentsCurrent map[string]*orbiter.NodeAgentCurrent) (orbiter.EnsureFunc, interface{}, interface{}, error) {

		desired := struct {
			Spec struct {
				Orbiter string
				Boom    string
			}
			Deps map[string][]byte
		}{}

	}
}
