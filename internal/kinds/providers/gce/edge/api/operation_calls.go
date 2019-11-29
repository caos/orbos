//go:generate stringer -type=Action

package api

import (
	"reflect"
	"time"

	"github.com/caos/orbiter/internal/core/logging"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
)

type Action int

const (
	Insert Action = iota
	Delete
	Add
	Remove
)

type GceCall interface {
	Do(opts ...googleapi.CallOption) (*compute.Operation, error)
}

type Operation struct {
	Logger       logging.Logger
	Action       Action
	ResourceType string
	ResourceName string
	GceCall      GceCall
}

func (c *Caller) RunFirstSuccessful(logger logging.Logger, action Action, gceCall ...GceCall) (compOp *compute.Operation, err error) {

	fieldedLogger := logger.WithFields(map[string]interface{}{"action": action})

next:
	for _, call := range gceCall {
		id := uuid.NewV1()

		callValue := reflect.ValueOf(call)
		c.addContext(callValue)
		callValue.MethodByName("RequestId").Call([]reflect.Value{reflect.ValueOf(id.String())})

		compOp, err = call.Do()
		if err != nil {
			continue next
		}

		for compOp.Progress < 100 {
			time.Sleep(time.Second)
			compOp, err = call.Do()
			if err != nil {
				continue next
			}
		}
		fieldedLogger.Info("Operation done")
	}
	return
}

/*
func (c *Caller) RunParallel(operation ...*Operation) ([]*compute.Operation, error) {

	compOps := make(chan *compute.Operation)
	errs := make(chan error)

	var wg sync.WaitGroup
	for _, op := range operation {
		wg.Add(1)
		go func(operation *Operation) {
			defer wg.Done()
			id := uuid.NewV1()

			callValue := reflect.ValueOf(operation.GceCall)
			c.addContext(callValue)
			callValue.MethodByName("RequestId").Call([]reflect.Value{reflect.ValueOf(id.String())})

			logger := operation.Logger.WithFields(map[string]interface{}{
				"action": operation.Action,
				"type":   operation.ResourceType,
				"name":   operation.ResourceName,
			})

			select {
			case <-errs:
				logger.Debug("Operation aborting")
				return
			default:
				logger.Debug("Operation starting")
			}

			op, err := operation.GceCall.Do()
			if err != nil {
				errs <- err
				return
			}

			for op.Progress < 100 {
				time.Sleep(time.Second)
				select {
				case <-errs:
					logger.Debug("Not waiting for operation to finish")
					return
				default:
				}
				op, err = operation.GceCall.Do()
				if err != nil {
					errs <- err
					return
				}
			}
			logger.Info("Operation done")

			compOps <- op
		}(op)
	}

	wg.Wait()
	close(errs)
	close(compOps)

	select {
	case err := <-errs:
		return nil, err
	default:
	}

	computeOperations := make([]*compute.Operation, 0)
	for compOp := range compOps {
		computeOperations = append(computeOperations, compOp)
	}
	return computeOperations, nil
}
*/
