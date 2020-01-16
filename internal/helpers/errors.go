package helpers

import (
	"fmt"
	"sync"
)

func Concat(left error, right error) error {
	if left == nil {
		return right
	}

	if right == nil {
		return left
	}

	return fmt.Errorf("%s | %s", right.Error(), left.Error())
}

// Synchronizer implements the error interface as well as
// the Causer interface from package github.com/pkg/errors
// It just returns the first error when Cause is called
//
// It is well suited for scatter and gather synchronization
type Synchronizer struct {
	errors []error
	wg     *sync.WaitGroup
	sync.RWMutex
}

func (s *Synchronizer) Cause() error {
	return s.errors[0]
}

func NewSynchronizer(wg *sync.WaitGroup) *Synchronizer {
	return &Synchronizer{errors: make([]error, 0), wg: wg}
}

func (s *Synchronizer) IsError() bool {
	s.Lock()
	defer s.Unlock()
	return len(s.errors) > 0
}

func (s *Synchronizer) Done(err error) {
	defer s.wg.Done()
	s.Lock()
	defer s.Unlock()
	if err == nil {
		return
	}
	s.errors = append(s.errors, err)
}

func (s Synchronizer) Error() string {
	built := ""
	for _, err := range s.errors {
		built = built + ", " + err.Error()
	}
	if len(built) >= 2 {
		built = built[2:]
	}
	return built
}
