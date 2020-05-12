package gce

import (
	"reflect"
	"testing"

	"github.com/caos/orbiter/internal/operator/orbiter"

	"github.com/caos/orbiter/internal/operator/orbiter/kinds/loadbalancers/dynamic"
)

func Test_normalize(t *testing.T) {
	type args struct {
		spec       map[string][]*dynamic.VIP
		orbID      string
		providerID string
	}
	tests := []struct {
		name string
		args args
		want []*normalizedLoadbalancer
	}{{
		name: "It should normalize correctly",
		args: args{
			spec: map[string][]*dynamic.VIP{
				"pool1": {{
					IP: strPtr("10.0.0.10"),
					Transport: []*dynamic.Source{{
						Name: "transport1",
						Destinations: []*dynamic.Destination{{
							HealthChecks: dynamic.HealthChecks{
								Protocol: "http",
								Path:     "/health",
								Code:     200,
							},
							Port: 30000,
							Pool: "target1",
						}},
						Whitelist: []*orbiter.CIDR{
							cidrPtr("0.0.0.0/0"),
						},
					}},
				}},
				"pool2": {{
					IP: strPtr("10.0.0.11"),
					Transport: []*dynamic.Source{{
						Name: "transport2",
						Destinations: []*dynamic.Destination{{
							HealthChecks: dynamic.HealthChecks{
								Protocol: "http",
								Path:     "/health",
								Code:     200,
							},
							Port: 30001,
							Pool: "target2",
						}},
						Whitelist: []*orbiter.CIDR{
							cidrPtr("0.0.0.0/0"),
						},
					}},
				}, {
					IP: strPtr("10.0.0.12"),
					Transport: []*dynamic.Source{{
						Name: "transport3",
						Destinations: []*dynamic.Destination{{
							HealthChecks: dynamic.HealthChecks{
								Protocol: "http",
								Path:     "/health",
								Code:     200,
							},
							Port: 30002,
							Pool: "target3",
						}},
						Whitelist: []*orbiter.CIDR{
							cidrPtr("0.0.0.0/0"),
						},
					}, {
						Name: "transport4",
						Destinations: []*dynamic.Destination{{
							HealthChecks: dynamic.HealthChecks{
								Protocol: "http",
								Path:     "/health",
								Code:     200,
							},
							Port: 30003,
							Pool: "target4",
						}, {
							HealthChecks: dynamic.HealthChecks{
								Protocol: "http",
								Path:     "/health",
								Code:     200,
							},
							Port: 30004,
							Pool: "target5",
						}},
						Whitelist: []*orbiter.CIDR{
							cidrPtr("0.0.0.0/0"),
						},
					}},
				}},
			},
			orbID:      "dummyorb",
			providerID: "dummyprovider",
		},
		want: []*normalizedLoadbalancer{{
			ip:             "10.0.0.10",
			forwardingRule: &forwardingRule{},
			targetPools:    []*targetPool{},
		}, {
			ip:             "10.0.0.11",
			forwardingRule: &forwardingRule{},
			targetPools:    []*targetPool{},
		}, {
			ip:             "10.0.0.12",
			forwardingRule: &forwardingRule{},
			targetPools:    []*targetPool{},
		}},
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalize(tt.args.spec, tt.args.orbID, tt.args.providerID); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("normalize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func strPtr(str string) *string {
	return &str
}

func cidrPtr(str string) *orbiter.CIDR {
	cidr := orbiter.CIDR(str)
	return &cidr
}
