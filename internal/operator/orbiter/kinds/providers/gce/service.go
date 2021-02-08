package gce

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"

	"google.golang.org/api/servicemanagement/v1"

	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

type svcConfig struct {
	monitor       mntr.Monitor
	networkName   string
	networkURL    string
	orbID         string
	providerID    string
	projectID     string
	desired       *Spec
	computeClient *compute.Service
	apiClient     *servicemanagement.APIService
	ctx           context.Context
	clientOptions *option.ClientOption
}

func service(ctx context.Context, monitor mntr.Monitor, desired *Spec, orbID, providerID string, oneoff bool) (*machinesService, func(), error) {

	jsonKey := []byte(desired.JSONKey.Value)
	/*
		c := &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
					DualStack: true,
				}).DialContext,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		}
	*/
	// avoids goroutine leaking
	// https://github.com/googleapis/google-cloud-go/issues/1025
	//	close := c.Transport.(*http.Transport).CloseIdleConnections

	close := func() {}
	clientOpts := []option.ClientOption{
		// option.WithHTTPClient(c),
		option.WithCredentialsJSON(jsonKey),
	}

	computeClient, err := compute.NewService(ctx, clientOpts...)
	if err != nil {
		return nil, close, err
	}

	apiClient, err := servicemanagement.NewService(ctx, clientOpts...)
	if err != nil {
		return nil, close, err
	}

	key := struct {
		ProjectID string `json:"project_id"`
	}{}
	if err := errors.Wrap(json.Unmarshal(jsonKey, &key), "extracting project id from jsonkey failed"); err != nil {
		return nil, close, err
	}

	monitor = monitor.WithField("projectID", key.ProjectID)
	h := fnv.New32()
	h.Write([]byte(orbID))
	networkName := fmt.Sprintf("orbos-network-%d", h.Sum32())
	h.Reset()
	networkURL := fmt.Sprintf("projects/%s/global/networks/%s", key.ProjectID, networkName)

	return newMachinesService(&svcConfig{
			monitor:       monitor,
			orbID:         orbID,
			providerID:    providerID,
			projectID:     key.ProjectID,
			desired:       desired,
			computeClient: computeClient,
			apiClient:     apiClient,
			ctx:           ctx,
			networkName:   networkName,
			networkURL:    networkURL,
		}, oneoff),
		close,
		nil
}
