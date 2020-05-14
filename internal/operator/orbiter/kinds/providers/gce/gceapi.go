package gce

import (
	"time"

	"google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
)

func toErrFunc(fn func()) func() error {
	return func() error {
		fn()
		return nil
	}
}

func operateFunc(before func(), call func(...googleapi.CallOption) (*compute.Operation, error), after func() error) func() error {
	return func() error {
		if before != nil {
			before()
		}
		op, err := call()
		if err != nil {
			return err
		}

		if op.Progress < 100 {
			time.Sleep(time.Second)
			return operateFunc(before, call, after)()
		}

		if after != nil {
			return after()
		}

		return nil
	}
}
