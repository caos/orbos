package providers

import (
	"regexp"
	"strings"

	"github.com/caos/orbiter/internal/operator/orbiter"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/gce"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/static"
	"github.com/caos/orbiter/internal/orb"
	"github.com/caos/orbiter/internal/secret"
	"github.com/caos/orbiter/internal/tree"
	"github.com/caos/orbiter/mntr"
	"github.com/pkg/errors"
)

var alphanum = regexp.MustCompile("[^a-zA-Z0-9]+")

func GetQueryAndDestroyFuncs(
	monitor mntr.Monitor,
	orb *orb.Orb,
	provID string,
	providerTree *tree.Tree,
	providerCurrent *tree.Tree,
	whitelistChan chan []*orbiter.CIDR,
) (
	orbiter.QueryFunc,
	orbiter.DestroyFunc,
	bool,
	error,
) {

	monitor = monitor.WithFields(map[string]interface{}{"provider": provID})

	wlFunc := func() []*orbiter.CIDR {
		monitor.Debug("Reading whitelist")
		return <-whitelistChan
	}

	switch providerTree.Common.Kind {
	case "orbiter.caos.ch/GCEProvider":
		return gce.AdaptFunc(
			orb.Masterkey,
			provID,
			alphanum.ReplaceAllString(strings.TrimSuffix(strings.TrimPrefix(orb.URL, "git@"), ".git"), "-"),
			wlFunc,
		)(
			monitor,
			providerTree,
			providerCurrent,
		)
	case "orbiter.caos.ch/StaticProvider":
		return static.AdaptFunc(
			orb.Masterkey,
			provID,
			wlFunc,
		)(
			monitor,
			providerTree,
			providerCurrent)
	default:
		return nil, nil, false, errors.Errorf("unknown provider kind %s", providerTree.Common.Kind)
	}

}

func GetSecrets(
	monitor mntr.Monitor,
	masterkey string,
	providerTree *tree.Tree,
) (
	map[string]*secret.Secret,
	error,
) {
	switch providerTree.Common.Kind {
	case "orbiter.caos.ch/GCEProvider":
		return gce.SecretsFunc(
			masterkey,
		)(
			monitor,
			providerTree,
		)
	case "orbiter.caos.ch/StaticProvider":
		return static.SecretsFunc(
			masterkey,
		)(
			monitor,
			providerTree,
		)
	default:
		return nil, errors.Errorf("unknown provider kind %s", providerTree.Common.Kind)
	}
}
