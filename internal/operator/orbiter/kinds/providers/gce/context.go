package gce

import (
	ctxpkg "context"
	"encoding/json"

	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

type context struct {
	monitor         mntr.Monitor
	orbID           string
	providerID      string
	projectID       string
	desired         *Spec
	client          *compute.Service
	machinesService *machinesService
	ctx             ctxpkg.Context
	auth            *option.ClientOption
}

func buildContext(monitor mntr.Monitor, desired *Spec, orbID, providerID string) (*context, error) {

	jsonKey := []byte(desired.JSONKey.Value)
	ctx := ctxpkg.Background()
	opt := option.WithCredentialsJSON(jsonKey)
	computeClient, err := compute.NewService(ctx, opt)
	if err != nil {
		return nil, err
	}

	key := struct {
		ProjectID string `json:"project_id"`
	}{}
	if err := errors.Wrap(json.Unmarshal(jsonKey, &key), "extracting project id from jsonkey failed"); err != nil {
		return nil, err
	}

	monitor = monitor.WithField("projectID", key.ProjectID)

	newContext := &context{
		monitor:    monitor,
		orbID:      orbID,
		providerID: providerID,
		projectID:  key.ProjectID,
		desired:    desired,
		client:     computeClient,
		ctx:        ctx,
		auth:       &opt,
	}
	newContext.machinesService = newMachinesService(newContext)
	return newContext, nil
}
