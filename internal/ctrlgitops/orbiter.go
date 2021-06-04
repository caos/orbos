package ctrlgitops

import (
	"context"
	"runtime/debug"
	"time"

	orbcfg "github.com/caos/orbos/pkg/orb"

	"github.com/caos/orbos/pkg/labels"

	"github.com/caos/orbos/internal/executables"
	"github.com/caos/orbos/internal/ingestion"
	"github.com/caos/orbos/internal/operator/orbiter"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/orb"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/git"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
)

type OrbiterConfig struct {
	Recur            bool
	Destroy          bool
	Deploy           bool
	Verbose          bool
	Version          string
	OrbConfigPath    string
	GitCommit        string
	IngestionAddress string
}

func Orbiter(ctx context.Context, monitor mntr.Monitor, conf *OrbiterConfig, orbctlGit *git.Client) error {

	go checks(monitor, orbctlGit)

	finishedChan := make(chan struct{})
	takeoffChan := make(chan struct{})
	healthyChan := make(chan bool)

	if conf.Recur {
		go orbiter.Instrument(monitor, healthyChan)
	} else {
		go func() {
			for range healthyChan {
			}
		}()
	}

	on := func() { takeoffChan <- struct{}{} }
	go on()
	var initialized bool
loop:
	for {
		select {
		case <-finishedChan:
			break loop
		case <-takeoffChan:
			iterate(conf, orbctlGit, !initialized, ctx, monitor, finishedChan, healthyChan, func(iterated bool) {
				if iterated {
					initialized = true
				}

				time.Sleep(time.Second * 30)
				go on()
			})
		}
	}
	return nil
}

func iterate(conf *OrbiterConfig, gitClient *git.Client, firstIteration bool, ctx context.Context, monitor mntr.Monitor, finishedChan chan struct{}, healthyChan chan bool, done func(iterated bool)) {

	var err error
	defer func() {
		go func() {
			if err != nil {
				healthyChan <- false
				return
			}
		}()
	}()

	orbFile, err := orbcfg.ParseOrbConfig(conf.OrbConfigPath)
	if err != nil {
		monitor.Error(err)
		done(false)
		return
	}

	if err := gitClient.Configure(orbFile.URL, []byte(orbFile.Repokey)); err != nil {
		monitor.Error(err)
		done(false)
		return
	}

	if err := gitClient.Clone(); err != nil {
		monitor.Error(err)
		done(false)
		return
	}

	pushEvents := func(events []*ingestion.EventRequest) error { return nil }
	if conf.IngestionAddress != "" {
		conn, err := grpc.Dial(conf.IngestionAddress, grpc.WithInsecure())
		if err != nil {
			panic(err)
		}

		ingc := ingestion.NewIngestionServiceClient(conn)
		defer conn.Close()

		pushEvents = func(events []*ingestion.EventRequest) error {
			_, err := ingc.PushEvents(ctx, &ingestion.EventsRequest{
				Orb:    orbFile.URL,
				Events: events,
			})
			return err
		}
	}

	if firstIteration {

		if err := pushEvents([]*ingestion.EventRequest{{
			CreationDate: ptypes.TimestampNow(),
			Type:         "orbiter.tookoff",
			Data: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"commit": {Kind: &structpb.Value_StringValue{StringValue: conf.GitCommit}},
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
							"since": {Kind: &structpb.Value_NumberValue{NumberValue: started}},
						},
					},
				}})
			}
		}()

		executables.Populate()

		monitor.WithFields(map[string]interface{}{
			"version": conf.Version,
			"commit":  conf.GitCommit,
			"destroy": conf.Destroy,
			"verbose": conf.Verbose,
			"repoURL": orbFile.URL,
		}).Info("Orbiter took off")
	}

	adaptFunc := orb.AdaptFunc(
		labels.MustForOperator("ORBOS", "orbiter.caos.ch", conf.Version),
		orbFile,
		conf.GitCommit,
		!conf.Recur,
		conf.Deploy,
		gitClient,
	)

	takeoffConf := &orbiter.Config{
		OrbiterCommit: conf.GitCommit,
		GitClient:     gitClient,
		Adapt:         adaptFunc,
		FinishedChan:  finishedChan,
		PushEvents:    pushEvents,
		OrbConfig:     *orbFile,
	}

	takeoff := orbiter.Takeoff(monitor, takeoffConf, healthyChan)

	go func() {
		started := time.Now()
		takeoff()

		monitor.WithFields(map[string]interface{}{
			"took": time.Since(started),
		}).Info("Iteration done")
		debug.FreeOSMemory()
		done(true)
	}()
}
