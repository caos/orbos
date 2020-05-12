package gce

import (
	"time"

	"google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
)

func operate(beforeRetry func(), call func(...googleapi.CallOption) (*compute.Operation, error)) error {
	op, err := call()
	if err != nil {
		return err
	}

	if op.Progress < 100 {
		beforeRetry()
		time.Sleep(time.Second)
		return operate(beforeRetry, call)
	}
	return nil
}
