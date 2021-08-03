package gce

import (
	"google.golang.org/api/googleapi"
	"google.golang.org/api/servicemanagement/v1"
)

func ensureIdentityAwareProxyAPIEnabled(c *context) error {
	_, err := servicemanagement.NewService(c.ctx, *c.auth)
	if err != nil {
		return err
	}

	return operateFunc(
		func() {
			c.monitor.Debug("Enabling Identity Aware Proxy API")
		},

		servicesOpCall(create),
		/*svc.Services.Create(&servicemanagement.ManagedService{
			ServiceName:       "iap.googleapis.com",
			ProducerProjectId: c.projectID,
		}).Do*/
		func() error {
			c.monitor.Debug("Identity Aware Proxy API ensured")
			return nil
		},
	)()
}

func create(opts ...googleapi.CallOption) (*servicemanagement.Operation, error) {
	return nil, nil
}
