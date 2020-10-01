package databases

import (
	"github.com/caos/orbos/internal/api"
	"github.com/caos/orbos/internal/operator/database/kinds/databases/core"
	orbdb "github.com/caos/orbos/internal/operator/database/kinds/orb"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/git"
	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/tree"
)

func Restore(
	monitor mntr.Monitor,
	k8sClient *kubernetes.Client,
	gitClient *git.Client,
	name string,
	databases []string,
) error {
	desired, err := api.ReadDatabaseYml(gitClient)
	if err != nil {
		monitor.Error(err)
		return err
	}
	current := &tree.Tree{}

	query, _, err := orbdb.AdaptFunc(name, "restore")(monitor, desired, current)
	if err != nil {
		monitor.Error(err)
		return err
	}
	queried := map[string]interface{}{}
	core.SetQueriedForDatabaseDBList(queried, databases)

	ensure, err := query(k8sClient, queried)
	if err != nil {
		monitor.Error(err)
		return err
	}

	if err := ensure(k8sClient); err != nil {
		monitor.Error(err)
		return err
	}
	return nil
}
