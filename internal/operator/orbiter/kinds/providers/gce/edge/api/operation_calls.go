//go:generate stringer -type=Action

package api

import (
	"reflect"
	"time"

	"github.com/caos/orbos/mntr"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/machine/v1"
)

type Action int

const (
	Insert Action = iota
	Delete
	Add
	Remove
)

type GceCall interface {
	Do(opts ...googleapi.CallOption) (*machine.Operation, error)
}

type Operation struct {
	monitor      mntr.Monitor
	Action       Action
	ResourceType string
	ResourceName string
	GceCall      GceCall
}

func (c *Caller) RunFirstSuccessful(monitor mntr.Monitor, action Action, gceCall ...GceCall) (compOp *machine.Operation, err error) {

	fieldedmonitor := monitor.WithFields(map[string]interface{}{"action": action})

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
		fieldedmonitor.Info("Operation done")
	}
	return
}

/*
func (c *Caller) RunParallel(operation ...*Operation) ([]*machine.Operation, error) {

	compOps := make(chan *machine.Operation)
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

			monitor := operation.monitor.WithFields(map[string]interface{}{
				"action": operation.Action,
				"type":   operation.ResourceType,
				"name":   operation.ResourceName,
			})

			select {
			case <-errs:
				monitor.Debug("Operation aborting")
				return
			default:
				monitor.Debug("Operation starting")
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
					monitor.Debug("Not waiting for operation to finish")
					return
				default:
				}
				op, err = operation.GceCall.Do()
				if err != nil {
					errs <- err
					return
				}
			}
			monitor.Info("Operation done")

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

	machineOperations := make([]*machine.Operation, 0)
	for compOp := range compOps {
		machineOperations = append(machineOperations, compOp)
	}
	return machineOperations, nil
}
*/
