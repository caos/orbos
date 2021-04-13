package helper

import "testing"

func TestIsNil(t *testing.T) {
	type args struct {
		sth interface{}
	}

	var unassigned *args
	var assigned *args = nil

	tests := []struct {
		name string
		args args
		want bool
	}{{
		name: "nil should be nil",
		args: args{sth: nil},
		want: true,
	}, {
		name: "unassigned should be nil",
		args: args{sth: unassigned},
		want: true,
	}, {
		name: "with nil assigned variable should be nil",
		args: args{sth: assigned},
		want: true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNil(tt.args.sth); got != tt.want {
				t.Errorf("IsNil() = %v, want %v", got, tt.want)
			}
		})
	}
}
