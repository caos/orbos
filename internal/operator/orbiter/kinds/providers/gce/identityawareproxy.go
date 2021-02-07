package gce

import (
	"fmt"

	"google.golang.org/api/servicemanagement/v1"
)

func ensureIdentityAwareProxyAPIEnabled(c *svcConfig) error {
	svc, err := servicemanagement.NewService(c.ctx, *c.auth)
	if err != nil {
		return err
	}

	return operateFunc(
		func() {
			c.monitor.Debug("Enabling Identity Aware Proxy API")
		},
		servicesOpCall(svc.Services.Enable("iap.googleapis.com", &servicemanagement.EnableServiceRequest{
			ConsumerId: fmt.Sprintf("project:%s", c.projectID),
		}).Do),
		func() error {
			c.monitor.Debug("Identity Aware Proxy API ensured")
			return nil
		},
	)()
}
