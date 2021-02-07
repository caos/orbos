package gce

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"

	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

type svcConfig struct {
	monitor     mntr.Monitor
	networkName string
	networkURL  string
	orbID       string
	providerID  string
	projectID   string
	desired     *Spec
	client      *compute.Service
	ctx         context.Context
	auth        *option.ClientOption
}

func service(ctx context.Context, monitor mntr.Monitor, desired *Spec, orbID, providerID string, oneoff bool) (*machinesService, error) {

	jsonKey := []byte(desired.JSONKey.Value)
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
	h := fnv.New32()
	h.Write([]byte(orbID))
	networkName := fmt.Sprintf("orbos-network-%d", h.Sum32())
	h.Reset()
	networkURL := fmt.Sprintf("projects/%s/global/networks/%s", key.ProjectID, networkName)

	return newMachinesService(&svcConfig{
		monitor:     monitor,
		orbID:       orbID,
		providerID:  providerID,
		projectID:   key.ProjectID,
		desired:     desired,
		client:      computeClient,
		ctx:         ctx,
		auth:        &opt,
		networkName: networkName,
		networkURL:  networkURL,
	}, oneoff), nil
}
