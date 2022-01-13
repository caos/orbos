package gce

import (
	"fmt"
	"strings"
	"time"

	"google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/servicemanagement/v1"
)

func toErrFunc(fn func()) func() error {
	return func() error {
		fn()
		return nil
	}
}

func computeOpCall(call func(...googleapi.CallOption) (*compute.Operation, error)) func() (*operation, error) {
	return func() (*operation, error) {
		op, err := call()
		if err != nil {
			return nil, err
		}
		if op != nil && op.Error != nil {
			return nil, opErrs(op.Error)
		}
		return toOperation(op.Progress), err
	}
}

type opErr struct {
	msg string
}

func (o *opErr) Error() string {
	return o.msg
}

func opErrs(errors *compute.OperationError) error {
	opErr := opErr{}
	for idx := range errors.Errors {
		opErr.msg = fmt.Sprintf("%s%s,", opErr.msg, errors.Errors[idx].Message)
	}
	opErr.msg = strings.TrimSuffix(opErr.msg, ",")
	return &opErr
}

func servicesOpCall(call func(...googleapi.CallOption) (*servicemanagement.Operation, error)) func() (*operation, error) {
	return func() (*operation, error) {
		_, err := call()
		return toOperation(100), err
	}
}

type operation struct {
	progress int64
}

func toOperation(progress int64) *operation {
	return &operation{progress: progress}
}

func operateFunc(before func(), call func() (*operation, error), after func() error) func() error {
	return func() error {
		if before != nil {
			before()
		}
		op, err := call()
		if err != nil {
			return err
		}

		if op.progress < 100 {
			time.Sleep(time.Second)
			return operateFunc(before, call, after)()
		}

		if after != nil {
			return after()
		}

		return nil
	}
}
