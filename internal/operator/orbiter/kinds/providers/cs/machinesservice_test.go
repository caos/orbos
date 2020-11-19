package cs

import (
	"testing"

	"github.com/cloudscale-ch/cloudscale-go-sdk"
)

type ifType string

var (
	private ifType = "private"
	public  ifType = "public"
)

func Test_createdIPs(t *testing.T) {

	newInterface := func(t ifType) cloudscale.Interface {
		return cloudscale.Interface{
			Type:    string(t),
			Network: cloudscale.NetworkStub{},
			Addresses: []cloudscale.Address{{
				Address: string(t),
			}},
		}
	}

	type args struct {
		interfaces []cloudscale.Interface
		oneoff     bool
	}
	tests := []struct {
		name  string
		args  args
		want  ifType
		want1 ifType
	}{{
		name: "sshing is done against public interface when in oneoff mode, for ORBITER, the private ip is relevant",
		args: args{
			interfaces: []cloudscale.Interface{
				newInterface(private),
				newInterface(public),
			},
			oneoff: true,
		},
		want:  private,
		want1: public,
	}, {
		name: "sshing is done against private interface when in recurring mode, for ORBITER, the private ip is relevant",
		args: args{
			interfaces: []cloudscale.Interface{
				newInterface(private),
				newInterface(public),
			},
			oneoff: false,
		},
		want:  private,
		want1: private,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := createdIPs(tt.args.interfaces, tt.args.oneoff)
			if got != string(tt.want) {
				t.Errorf("createdIPs() got = %v, want %v", got, tt.want)
			}
			if got1 != string(tt.want1) {
				t.Errorf("createdIPs() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
