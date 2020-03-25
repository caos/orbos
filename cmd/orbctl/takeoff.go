package main

import (
	"time"

	"github.com/golang/protobuf/ptypes"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"github.com/caos/orbiter/internal/executables"
	"github.com/caos/orbiter/internal/ingestion"
	"github.com/caos/orbiter/internal/operator"
	"github.com/caos/orbiter/internal/operator/orbiter"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/orb"
	"github.com/caos/orbiter/internal/watcher/cron"
	"github.com/caos/orbiter/internal/watcher/immediate"
)

func takeoffCommand(rv rootValues) *cobra.Command {

	var (
		verbose bool
		recur   bool
		destroy bool
		deploy  bool
		ingestionAddress string
		cmd     = &cobra.Command{
			Use:   "takeoff",
			Short: "Launch an orbiter",
			Long:  "Ensures a desired state",
		}
	)

	flags := cmd.Flags()
	flags.BoolVar(&recur, "recur", false, "Ensure the desired state continously")
	flags.BoolVar(&deploy, "deploy", true, "Ensure Orbiter and Boom deployments continously")
	flags.StringVar(&ingestionAddress, "ingestion", "", "Ingestion API address")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if recur && destroy {
			return errors.New("flags --recur and --destroy are mutually exclusive, please provide eighter one or none")
		}

		ctx, monitor, gitClient, orbFile, errFunc := rv()
		if errFunc != nil {
			return errFunc(cmd)
		}

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

		op := operator.New(ctx, monitor, orbiter.Takeoff(
			monitor,
			gitClient,
			pushEvents,
			gitCommit,
			orb.AdaptFunc(
				orbFile,
				gitCommit,
				!recur,
				deploy),
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

		op.Run()

		return nil
	}
	return cmd
}
