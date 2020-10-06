package cs

import (
	"testing"

	"github.com/cloudscale-ch/cloudscale-go-sdk"
)

func Test_createdIPs(t *testing.T) {
	type args struct {
		interfaces []cloudscale.Interface
		oneoff     bool
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := createdIPs(tt.args.interfaces, tt.args.oneoff)
			if got != tt.want {
				t.Errorf("createdIPs() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("createdIPs() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
