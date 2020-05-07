package start

import (
	"context"
	"github.com/caos/orbos/internal/executables"
	"github.com/caos/orbos/internal/git"
	"github.com/caos/orbos/internal/ingestion"
	"github.com/caos/orbos/internal/operator"
	"github.com/caos/orbos/internal/operator/boom"
	"github.com/caos/orbos/internal/operator/orbiter"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/orb"
	orbconfig "github.com/caos/orbos/internal/orb"
	"github.com/caos/orbos/internal/watcher/cron"
	"github.com/caos/orbos/internal/watcher/immediate"
	"github.com/caos/orbos/mntr"
	"github.com/golang/protobuf/ptypes"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"google.golang.org/grpc"
	"time"
)

func Orbiter(ctx context.Context, monitor mntr.Monitor, recur, destroy, deploy, verbose bool, version string, gitClient *git.Client, orbFile *orbconfig.Orb, gitCommit string, ingestionAddress string) error {

	finishedChan := make(chan bool)
	ctxCancel, cancel := context.WithCancel(ctx)

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
			_, err := ingc.PushEvents(ctxCancel, &ingestion.EventsRequest{
				Orb:    orbFile.URL,
				Events: events,
			})
			return err
		}
	}

	op := operator.New(ctxCancel, monitor, orbiter.Takeoff(
		monitor,
		gitClient,
		pushEvents,
		gitCommit,
		orb.AdaptFunc(
			orbFile,
			gitCommit,
			!recur,
			deploy),
		finishedChan,
	), []operator.Watcher{
		immediate.New(monitor),
		cron.New(monitor, "@every 10s"),
	})

	if err := op.Initialize(); err != nil {
		panic(err)
	}

	executables.Populate()

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

	monitor.WithFields(map[string]interface{}{
		"version": version,
		"commit":  gitCommit,
		"destroy": destroy,
		"verbose": verbose,
		"repoURL": orbFile.URL,
	}).Info("Orbiter took off")

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

	go func() {
		op.Run()
	}()

	finished := false
	for !finished {
		finished = <-finishedChan
	}
	cancel()

	return nil
}

func Boom(ctx context.Context, monitor mntr.Monitor, orbFile *orbconfig.Orb, localmode bool) error {

	finishedChan := make(chan bool)
	ctxCancel, cancel := context.WithCancel(ctx)

	op := operator.New(ctxCancel, monitor, boom.Takeoff(
		monitor,
		orbFile,
		"/boom",
		localmode,
		finishedChan,
	), []operator.Watcher{
		immediate.New(monitor),
		cron.New(monitor, "@every 60s"),
	})

	if err := op.Initialize(); err != nil {
		panic(err)
	}

	go func() {
		op.Run()
	}()

	opMetrics := operator.New(ctxCancel, monitor, boom.TakeOffCurrentState(
		monitor,
		orbFile,
		"/boom",
		finishedChan,
	), []operator.Watcher{
		immediate.New(monitor),
		cron.New(monitor, "@every 60s"),
	})

	if err := opMetrics.Initialize(); err != nil {
		panic(err)
	}

	go func() {
		opMetrics.Run()
	}()

	finished := false
	for !finished {
		finished = <-finishedChan
	}
	cancel()

	return nil
}
