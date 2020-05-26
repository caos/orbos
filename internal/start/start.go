package start

import (
	"context"
	"errors"
	"github.com/caos/orbos/internal/executables"
	"github.com/caos/orbos/internal/git"
	"github.com/caos/orbos/internal/ingestion"
	"github.com/caos/orbos/internal/operator/boom"
	"github.com/caos/orbos/internal/operator/orbiter"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/orb"
	"github.com/caos/orbos/internal/operator/secretfuncs"
	orbconfig "github.com/caos/orbos/internal/orb"
	"github.com/caos/orbos/internal/secret"
	"github.com/caos/orbos/mntr"
	"github.com/golang/protobuf/ptypes"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"google.golang.org/grpc"
	"runtime/debug"
	"strings"
	"time"
)

func Orbiter(ctx context.Context, monitor mntr.Monitor, recur, destroy, deploy, verbose bool, version string, gitClient *git.Client, orbFile *orbconfig.Orb, gitCommit string, ingestionAddress string) ([]string, error) {
	orbiter.Metrics()

	finishedChan := make(chan bool)

	pushEvents := func(_ []*ingestion.EventRequest) error {
		return nil
	}
	if ingestionAddress != "" {
		conn, err := grpc.Dial(ingestionAddress, grpc.WithInsecure())
		if err != nil {
			panic(err)
		}

		ingc := ingestion.NewIngestionServiceClient(conn)

		pushEvents = func(events []*ingestion.EventRequest) error {
			_, err := ingc.PushEvents(ctx, &ingestion.EventsRequest{
				Orb:    orbFile.URL,
				Events: events,
			})
			return err
		}
	}

	if err := pushEvents([]*ingestion.EventRequest{{
		CreationDate: ptypes.TimestampNow(),
		Type:         "orbiter.tookoff",
		Data: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"commit": &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: gitCommit}},
			},
		},
	}}); err != nil {
		panic(err)
	}

	started := float64(time.Now().UTC().Unix())

	go func() {
		for range time.Tick(time.Minute) {
			pushEvents([]*ingestion.EventRequest{{
				CreationDate: ptypes.TimestampNow(),
				Type:         "orbiter.running",
				Data: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"since": &structpb.Value{Kind: &structpb.Value_NumberValue{NumberValue: started}},
					},
				},
			}})
		}
	}()

	executables.Populate()

	monitor.WithFields(map[string]interface{}{
		"version": version,
		"commit":  gitCommit,
		"destroy": destroy,
		"verbose": verbose,
		"repoURL": orbFile.URL,
	}).Info("Orbiter took off")

	go func() {
		takeoffChan := make(chan struct{})
		go func() {
			takeoffChan <- struct{}{}
		}()

		for range takeoffChan {
			adaptFunc := orb.AdaptFunc(
				orbFile,
				gitCommit,
				!recur,
				deploy)

			takeoff := orbiter.Takeoff(
				monitor,
				gitClient,
				pushEvents,
				gitCommit,
				adaptFunc,
				finishedChan,
			)

			go func() {
				started := time.Now()
				takeoff()

				monitor.WithFields(map[string]interface{}{
					"took": time.Since(started),
				}).Info("Iteration done")
				debug.FreeOSMemory()
				takeoffChan <- struct{}{}
			}()
		}
	}()

	finished := false
	for !finished {
		finished = <-finishedChan
	}

	kubeconfigs := make([]string, 0)

	orbTree, err := orbiter.Parse(gitClient, "orbiter.yml")
	if err != nil {
		return nil, errors.New("Failed to parse orbiter.yml")
	}

	orbDef, err := orb.ParseDesiredV0(orbTree[0])
	if err != nil {
		return nil, errors.New("Failed to parse orbiter.yml")
	}

	for clustername, _ := range orbDef.Clusters {
		path := strings.Join([]string{"orbiter", clustername, "kubeconfig"}, ".")

		value, err := secret.Read(
			monitor,
			gitClient,
			secretfuncs.Get(orbFile),
			path)
		if err != nil || value == "" {
			return nil, errors.New("Failed to get kubeconfig")
		}
		monitor.Info("Read kubeconfig for boom deployment")

		kubeconfigs = append(kubeconfigs, value)
	}

	return kubeconfigs, nil
}

func Boom(monitor mntr.Monitor, orbFile *orbconfig.Orb, localmode bool) error {
	boom.Metrics(monitor)

	takeoffChan := make(chan struct{})
	go func() {
		takeoffChan <- struct{}{}
	}()

	for range takeoffChan {
		boomChan := make(chan struct{})
		currentChan := make(chan struct{})

		takeoffCurrent := boom.TakeOffCurrentState(
			monitor,
			orbFile,
			"/boom",
		)
		go func() {
			started := time.Now()
			takeoffCurrent()

			monitor.WithFields(map[string]interface{}{
				"took": time.Since(started),
			}).Info("Iteration done")
			debug.FreeOSMemory()

			currentChan <- struct{}{}
		}()

		takeoff := boom.Takeoff(
			monitor,
			orbFile,
			"/boom",
			localmode,
		)
		go func() {
			started := time.Now()
			takeoff()

			monitor.WithFields(map[string]interface{}{
				"took": time.Since(started),
			}).Info("Iteration done")
			debug.FreeOSMemory()

			boomChan <- struct{}{}
		}()

		go func() {
			<-currentChan
			<-boomChan

			takeoffChan <- struct{}{}
		}()
	}

	return nil
}
