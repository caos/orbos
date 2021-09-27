package orbiter

import (
	"github.com/caos/orbos/v5/pkg/git"
	orbcfg "github.com/caos/orbos/v5/pkg/orb"
)

type Config struct {
	OrbiterCommit string
	GitClient     *git.Client
	Adapt         AdaptFunc
	FinishedChan  chan struct{}
	OrbConfig     orbcfg.Orb
}
