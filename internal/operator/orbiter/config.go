package orbiter

import (
	"github.com/caos/orbos/internal/git"
	"github.com/caos/orbos/internal/ingestion"
)

type Config struct {
	OrbiterCommit string
	GitClient     *git.Client
	Adapt         AdaptFunc
	FinishedChan  chan bool
	PushEvents    func(events []*ingestion.EventRequest) error
}
