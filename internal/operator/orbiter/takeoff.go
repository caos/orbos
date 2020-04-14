package orbiter

import (
	"github.com/caos/orbiter/internal/push"
	"github.com/caos/orbiter/internal/tree"
	_ "net/http/pprof"

	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/yaml.v3"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/caos/orbiter/internal/git"
	"github.com/caos/orbiter/internal/ingestion"
	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/mntr"
)

type EnsureFunc func(psf push.Func) error

type QueryFunc func(nodeAgentsCurrent map[string]*common.NodeAgentCurrent, nodeAgentsDesired map[string]*common.NodeAgentSpec, queried map[string]interface{}) (EnsureFunc, error)

type event struct {
	commit string
	files  []git.File
}

func Takeoff(monitor mntr.Monitor, gitClient *git.Client, pushEvents func(events []*ingestion.EventRequest) error, orbiterCommit string, adapt AdaptFunc) func() {

	go func() {
		prometheus.MustRegister(prometheus.NewBuildInfoCollector())
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(":9000", nil); err != nil {
			panic(err)
		}
	}()

	return func() {

		trees, err := parse(gitClient, "orbiter.yml")
		if err != nil {
			monitor.Error(err)
			return
		}

		treeDesired := trees[0]
		treeCurrent := &tree.Tree{}

		desiredNodeAgents := common.NodeAgentsDesiredKind{
			Kind:    "nodeagent.caos.ch/NodeAgents",
			Version: "v0",
		}
		rawDesiredNodeAgents, err := gitClient.Read("caos-internal/orbiter/node-agents-desired.yml")
		if err != nil {
			panic(err)
		}
		yaml.Unmarshal(rawDesiredNodeAgents, &desiredNodeAgents)
		desiredNodeAgents.Kind = "nodeagent.caos.ch/NodeAgents"
		desiredNodeAgents.Version = "v0"
		desiredNodeAgents.Spec.Commit = orbiterCommit
		if desiredNodeAgents.Spec.NodeAgents == nil {
			desiredNodeAgents.Spec.NodeAgents = make(map[string]*common.NodeAgentSpec)
		}

		marshalCurrentFiles := func() []git.File {
			return []git.File{{
				Path:    "caos-internal/orbiter/current.yml",
				Content: common.MarshalYAML(treeCurrent),
			}, {
				Path:    "caos-internal/orbiter/node-agents-desired.yml",
				Content: common.MarshalYAML(desiredNodeAgents),
			}}
		}

		events := make([]*event, 0)
		monitor.OnChange = mntr.Concat(func(evt string, fields map[string]string) {
			pushEvents([]*ingestion.EventRequest{mntr.EventRecord("orbiter", evt, fields)})
			events = append(events, &event{
				commit: mntr.CommitRecord(mntr.AggregateCommitFields(fields)),
				files:  marshalCurrentFiles(),
			})
		}, monitor.OnChange)

		query, _, migrate, err := adapt(monitor, treeDesired, treeCurrent)
		if err != nil {
			monitor.Error(err)
			return
		}

		if migrate {
			if err := push.OrbiterYML(monitor, "Desired state migrated", gitClient, treeDesired); err != nil {
				monitor.Error(err)
				return
			}
		}

		currentNodeAgents := common.NodeAgentsCurrentKind{}
		rawCurrentNodeAgents, err := gitClient.Read("caos-internal/orbiter/node-agents-current.yml")
		if err != nil {
			panic(err)
		}
		yaml.Unmarshal(rawCurrentNodeAgents, &currentNodeAgents)

		if currentNodeAgents.Current == nil {
			currentNodeAgents.Current = make(map[string]*common.NodeAgentCurrent)
		}

		handleAdapterError := func(err error) {
			monitor.Error(err)
			//			monitor.Error(gitClient.Clone())
			if commitErr := gitClient.Commit(mntr.CommitRecord([]*mntr.Field{{Pos: 0, Key: "err", Value: err.Error()}})); commitErr != nil {
				panic(commitErr)
			}
			monitor.Error(gitClient.Push())
		}

		ensure, err := query(currentNodeAgents.Current, desiredNodeAgents.Spec.NodeAgents, nil)
		if err != nil {
			handleAdapterError(err)
			return
		}

		if err := gitClient.Clone(); err != nil {
			panic(err)
		}

		reconciledCurrentStateMsg := "Current state reconciled"
		currentReconciled, err := gitClient.StageAndCommit(mntr.CommitRecord([]*mntr.Field{{Key: "evt", Value: reconciledCurrentStateMsg}}), marshalCurrentFiles()...)
		if err != nil {
			panic(fmt.Errorf("Commiting event \"%s\" failed: %s", reconciledCurrentStateMsg, err.Error()))
		}

		if currentReconciled {
			if err := gitClient.Push(); err != nil {
				panic(err)
			}
		}

		events = make([]*event, 0)
		if err := ensure(push.SecretsFunc(gitClient, treeDesired)); err != nil {
			handleAdapterError(err)
			return
		}

		if err := gitClient.Clone(); err != nil {
			panic(err)
		}

		for _, event := range events {

			changed, err := gitClient.StageAndCommit(event.commit, event.files...)
			if err != nil {
				panic(fmt.Errorf("Commiting event failed with err %s: %s", err.Error(), event.commit))
			}

			if !changed {
				panic(fmt.Sprint("Event has no effect:", event.commit))
			}
		}

		if len(events) > 0 {
			monitor.Error(gitClient.Push())
		}
	}
}
