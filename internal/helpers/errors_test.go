package helpers_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/caos/orbos/v5/internal/helpers"
	"github.com/caos/orbos/v5/mntr"
)

func TestConcatKeepsUserErrorType(t *testing.T) {
	type args struct {
		left  error
		right error
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{{
		name: "When both errors are of type UserError, the returned error should also be of type UserError",
		args: args{
			left:  mntr.UserError{Err: errors.New("left")},
			right: mntr.UserError{Err: errors.New("right")},
		},
		wantErr: true,
	}, {
		name: "When only left error is of type UserError, the returned error should not be of type UserError",
		args: args{
			left:  mntr.UserError{Err: errors.New("left")},
			right: errors.New("right"),
		},
		wantErr: false,
	}, {
		name: "When only right error is of type UserError, the returned error should not be of type UserError",
		args: args{
			left:  errors.New("left"),
			right: mntr.UserError{Err: errors.New("right")},
		},
		wantErr: false,
	}, {
		name: "When only left error is non-nil and of type UserError, the returned error should also be of type UserError",
		args: args{
			left:  mntr.UserError{Err: errors.New("left")},
			right: nil,
		},
		wantErr: true,
	}, {
		name: "When only right error is non-nil and of type UserError, the returned error should also be of type UserError",
		args: args{
			left:  nil,
			right: mntr.UserError{Err: errors.New("right")},
		},
		wantErr: true,
	}, {
		name: "When only left error is non-nil but not of type UserError, the returned error should also not be of type UserError",
		args: args{
			left:  errors.New("left"),
			right: nil,
		},
		wantErr: false,
	}, {
		name: "When only right error is non-nil but not of type UserError, the returned error should also not be of type UserError",
		args: args{
			left:  nil,
			right: errors.New("right"),
		},
		wantErr: false,
	}, {
		name: "When both errors have a nested type UserError, the returned error should also be of type UserError",
		args: args{
			left:  fmt.Errorf("some UserError: %w", mntr.UserError{Err: errors.New("left")}),
			right: fmt.Errorf("second level error: %w", fmt.Errorf("first level error: %w", mntr.UserError{Err: errors.New("right")})),
		},
		wantErr: true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := helpers.Concat(tt.args.left, tt.args.right); errors.As(err, &mntr.UserError{}) != tt.wantErr {
				t.Errorf("Concat() error = %v, wantUserError %v", err, tt.wantErr)
			}
		})
	}
}
