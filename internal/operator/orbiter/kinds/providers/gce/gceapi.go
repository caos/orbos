package gce

import (
	"time"

	"google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
)

func operate(before func(), call func(...googleapi.CallOption) (*compute.Operation, error)) error {
	before()
	op, err := call()
	if err != nil {
		return err
	}

	if op.Progress < 100 {
		time.Sleep(time.Second)
		return operate(before, call)
	}
	return nil
}
