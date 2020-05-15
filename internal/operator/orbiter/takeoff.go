package orbiter

import (
	"fmt"
	"net/http"

	"github.com/caos/orbos/internal/push"
	"github.com/caos/orbos/internal/tree"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/yaml.v3"

	"github.com/caos/orbos/internal/git"
	"github.com/caos/orbos/internal/ingestion"
	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/mntr"
)

type EnsureFunc func(psf push.Func) error

type QueryFunc func(nodeAgentsCurrent map[string]*common.NodeAgentCurrent, nodeAgentsDesired map[string]*common.NodeAgentSpec, queried map[string]interface{}) (EnsureFunc, error)

type retQuery struct {
	ensure EnsureFunc
	err    error
}

func QueryFuncGoroutine(query func() (EnsureFunc, error)) (EnsureFunc, error) {
	retChan := make(chan retQuery)
	go func() {
		ensure, err := query()
		retChan <- retQuery{ensure, err}
	}()
	ret := <-retChan
	return ret.ensure, ret.err
}
func EnsureFuncGoroutine(ensure func() error) error {
	retChan := make(chan error)
	go func() {
		retChan <- ensure()
	}()
	return <-retChan
}

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
		rawDesiredNodeAgents := gitClient.Read("caos-internal/orbiter/node-agents-desired.yml")
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

		adaptFunc := func() (QueryFunc, DestroyFunc, bool, error) {
			return adapt(monitor, treeDesired, treeCurrent)
		}
		query, _, migrate, err := AdaptFuncGoroutine(adaptFunc)
		if err != nil {
			monitor.Error(err)
			return
		}

		if migrate {
			if err := push.YML(monitor, "Desired state migrated", gitClient, treeDesired, "orbiter.yml"); err != nil {
				monitor.Error(err)
				return
			}
		}

		currentNodeAgents := common.NodeAgentsCurrentKind{}
		yaml.Unmarshal(gitClient.Read("caos-internal/orbiter/node-agents-current.yml"), &currentNodeAgents)

		if currentNodeAgents.Current == nil {
			currentNodeAgents.Current = make(map[string]*common.NodeAgentCurrent)
		}

		handleAdapterError := func(err error) {
			monitor.Error(err)
			//			monitor.Error(gitClient.Clone())
			if commitErr := gitClient.Commit(mntr.CommitRecord([]*mntr.Field{{Pos: 0, Key: "err", Value: err.Error()}})); commitErr != nil {
				monitor.Error(err)
				return
			}
			monitor.Error(gitClient.Push())
		}

		queryFunc := func() (EnsureFunc, error) {
			return query(currentNodeAgents.Current, desiredNodeAgents.Spec.NodeAgents, nil)
		}
		ensure, err := QueryFuncGoroutine(queryFunc)
		if err != nil {
			handleAdapterError(err)
			return
		}

		if err := gitClient.Clone(); err != nil {
			monitor.Error(err)
			return
		}

		reconciledCurrentStateMsg := "Current state reconciled"
		currentReconciled, err := gitClient.StageAndCommit(mntr.CommitRecord([]*mntr.Field{{Key: "evt", Value: reconciledCurrentStateMsg}}), marshalCurrentFiles()...)
		if err != nil {
			monitor.Error(fmt.Errorf("Commiting event \"%s\" failed: %s", reconciledCurrentStateMsg, err.Error()))
			return
		}

		if currentReconciled {
			if err := gitClient.Push(); err != nil {
				monitor.Error(fmt.Errorf("Pushing event \"%s\" failed: %s", reconciledCurrentStateMsg, err.Error()))
				return
			}
		}

		events = make([]*event, 0)

		ensureFunc := func() error {
			return ensure(push.SecretsFunc(gitClient, treeDesired, "orbiter.yml"))
		}
		if err := EnsureFuncGoroutine(ensureFunc); err != nil {
			handleAdapterError(err)
			return
		}

		if err := gitClient.Clone(); err != nil {
			monitor.Error(fmt.Errorf("Commiting event \"%s\" failed: %s", reconciledCurrentStateMsg, err.Error()))
			return
		}

		for _, event := range events {

			changed, err := gitClient.StageAndCommit(event.commit, event.files...)
			if err != nil {
				monitor.Error(fmt.Errorf("Commiting event \"%s\" failed: %s", event.commit, err.Error()))
				return
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
