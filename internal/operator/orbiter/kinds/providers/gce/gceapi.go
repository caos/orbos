package gce

import (
	"time"

	"google.golang.org/api/servicemanagement/v1"

	"google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
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
		return toOperation(op.Progress), err
	}
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
