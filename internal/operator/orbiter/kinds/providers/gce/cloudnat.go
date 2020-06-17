package gce

import (
	"fmt"

	uuid "github.com/satori/go.uuid"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
)

const CLOUD_NAT_NAME = "orbos-cloud-nat"

func destroyCloudNAT(c *context) error {
	svc, err := compute.NewService(c.ctx, *c.auth)
	if err != nil {
		return err
	}

	if delErr, ok := operateFunc(
		func() {
			c.monitor.Debug("Deleting Cloud NAT Router")
		},
		computeOpCall(svc.Routers.Delete(c.projectID, c.desired.Region, CLOUD_NAT_NAME).RequestId(uuid.NewV1().String()).Do),
		func() error {
			c.monitor.Info("Cloud NAT Router deleted")
			return nil
		},
	)().(*googleapi.Error); ok && delErr.Code != 404 {
		return delErr
	}
	return nil
}

func ensureCloudNAT(c *context) error {
	svc, err := compute.NewService(c.ctx, *c.auth)
	if err != nil {
		return err
	}

	_, err = svc.Routers.Get(c.projectID, c.desired.Region, CLOUD_NAT_NAME).Do()
	if e, ok := err.(*googleapi.Error); ok && e.Code == 404 {
		return operateFunc(
			func() {
				c.monitor.Debug("Creating Cloud NAT Router")
			},
			computeOpCall(svc.Routers.Insert(c.projectID, c.desired.Region, &compute.Router{
				Name:    CLOUD_NAT_NAME,
				Network: fmt.Sprintf("projects/%s/global/networks/default", c.projectID),
				Nats: []*compute.RouterNat{{
					Name:                          CLOUD_NAT_NAME,
					NatIpAllocateOption:           "AUTO_ONLY",
					SourceSubnetworkIpRangesToNat: "ALL_SUBNETWORKS_ALL_IP_RANGES",
				}},
			}).RequestId(uuid.NewV1().String()).Do),
			func() error {
				c.monitor.Info("Cloud NAT Router created")
				return nil
			},
		)()
	}
	return err
}
