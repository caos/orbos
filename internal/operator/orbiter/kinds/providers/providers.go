package providers

import (
	"regexp"
	"strings"

	"github.com/caos/orbos/internal/operator/orbiter"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/providers/gce"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/providers/static"
	"github.com/caos/orbos/internal/orb"
	"github.com/caos/orbos/internal/secret"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
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
	finishedChan chan bool,
	orbiterCommit, repoURL, repoKey string,
	oneoff bool,
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
			orbiterCommit, repoURL, repoKey,
			oneoff,
		)(
			monitor,
			finishedChan,
			providerTree,
			providerCurrent,
		)
	case "orbiter.caos.ch/StaticProvider":

		adaptFunc := func() (orbiter.QueryFunc, orbiter.DestroyFunc, bool, error) {
			return static.AdaptFunc(
				orb.Masterkey,
				provID,
				func() []*orbiter.CIDR {
					monitor.Debug("Reading whitelist")
					return <-whitelistChan
				},
				orbiterCommit, repoURL, repoKey,
			)(
				monitor.WithFields(map[string]interface{}{"provider": provID}),
				finishedChan,
				providerTree,
				providerCurrent)
		}
		return orbiter.AdaptFuncGoroutine(adaptFunc)
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

func RewriteMasterkey(
	monitor mntr.Monitor,
	oldMasterkey string,
	newMasterkey string,
	providerTree *tree.Tree,
) (
	map[string]*secret.Secret,
	error,
) {
	switch providerTree.Common.Kind {
	case "orbiter.caos.ch/StaticProvider":
		return static.RewriteFunc(
			oldMasterkey,
			newMasterkey,
		)(
			monitor,
			providerTree,
		)
	default:
		return nil, errors.Errorf("unknown provider kind %s", providerTree.Common.Kind)
	}
}
