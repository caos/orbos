package core

import (
	"sync"

	"github.com/caos/orbiter/internal/core/helpers"
	"github.com/goombaio/dag"
)

type ResourceService interface {
	// Neighter include dependencies nor an ID in interface{}
	Abbreviate() string
	Desire(payload interface{}) (interface{}, error)
	Ensure(id string, desired interface{}, ensuredDependencies []interface{}) (interface{}, error)
	Cleanupper
}

type Cleanupper interface {
	AllExisting() ([]string, error)
	Delete(id string) error
}

type Executor struct {
	ensureGraph *dag.DAG
	cleanuppers [][]Cleanupper
	mux         sync.RWMutex
}

type vertex struct {
	ensuring chan struct{}
	ensured  interface{}
	resource *Resource
	mux      *sync.Mutex
}

// NewExecutor - inner cleanuppers are executed concurrently,
//               cleanupper slices are executed sequentially
func NewExecutor(listersInCleanupOrder [][]Cleanupper) *Executor {
	return &Executor{
		ensureGraph: dag.NewDAG(),
		cleanuppers: listersInCleanupOrder,
	}
}

func (e *Executor) Plan(root *Resource) error {
	_, err := e.planResource(root)
	return err
}

func (e *Executor) planResource(resource *Resource) (*dag.Vertex, error) {
	id, err := resource.ID()
	if err != nil {
		return nil, err
	}

	libVertex, err := e.ensureGraph.GetVertex(*id)
	if libVertex == nil && err != nil {
		libVertex = dag.NewVertex(*id, &vertex{
			resource: resource,
			mux:      &sync.Mutex{},
		})
		e.ensureGraph.AddVertex(libVertex)
	}

	for _, dep := range resource.dependencies {
		depVertex, err := e.planResource(dep)
		if err != nil {
			return nil, err
		}
		e.ensureGraph.AddEdge(libVertex, depVertex)
	}

	return libVertex, nil
}

func (e *Executor) Run() (chan error, error) {
	if err := e.ensure(); err != nil {
		return nil, err
	}

	cleanupped := make(chan error, 1)

	go e.cleanup(cleanupped)

	return cleanupped, nil
}

func (e *Executor) ensure() error {

	sources := e.ensureGraph.SourceVertices()

	var wg sync.WaitGroup
	synchronizer := helpers.NewSynchronizer(&wg)
	wg.Add(len(sources))
	for _, sourceVertex := range sources {
		go func(vertex *dag.Vertex) {
			_, err := e.ensureVertex(vertex)
			synchronizer.Done(err)
		}(sourceVertex)
	}

	wg.Wait()

	if synchronizer.IsError() {
		return synchronizer
	}

	return nil
}

func (e *Executor) ensureVertex(libVertex *dag.Vertex) (interface{}, error) {

	resourceVertex := libVertex.Value.(*vertex)
	resourceVertex.mux.Lock()
	defer resourceVertex.mux.Unlock()
	if resourceVertex.ensuring != nil || resourceVertex.ensured != nil {
	loop:
		for {
			select {
			case <-resourceVertex.ensuring:
				break loop
			}
		}
		if err, ok := resourceVertex.ensured.(error); ok {
			return nil, err
		}
		return resourceVertex.ensured, nil
	}

	resourceVertex.ensuring = make(chan struct{})
	defer close(resourceVertex.ensuring)

	children, err := e.ensureGraph.Successors(libVertex)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	synchronizer := helpers.NewSynchronizer(&wg)
	wg.Add(len(children))
	deps := make([]interface{}, 0)
	for _, child := range children {
		go func(v *dag.Vertex) {
			ensuredChild, goErr := e.ensureVertex(v)
			if ensuredChild != nil {
				e.mux.Lock()
				deps = append(deps, ensuredChild)
				e.mux.Unlock()
			}
			synchronizer.Done(goErr)
		}(child)
	}

	wg.Wait()
	if synchronizer.IsError() {
		return nil, synchronizer
	}

	id, err := resourceVertex.resource.ID()
	if err != nil {
		return nil, err
	}

	desired, err := resourceVertex.resource.desire()
	if err != nil {
		return nil, err
	}

	ret, err := resourceVertex.resource.service.Ensure(*id, desired, deps)
	if err != nil {
		return nil, err
	}

	resourceVertex.ensured = ret
	if resourceVertex.resource.ensured != nil {
		return ret, resourceVertex.resource.ensured(ret)
	}

	return ret, nil
}

func (e *Executor) cleanup(cleanupped chan<- error) {

	var mux sync.Mutex
	var err error

outer:
	for _, seq := range e.cleanuppers {
		var wg sync.WaitGroup
		for _, ccur := range seq {
			var all []string
			all, err = ccur.AllExisting()
			if err != nil {
				break outer
			}

			for _, existing := range all {
				wg.Add(1)
				go func(cleanupper Cleanupper, resource string) {
					defer wg.Done()
					if goErr := e.cleanupResource(cleanupper, resource); goErr != nil {
						mux.Lock()
						err = goErr
						mux.Unlock()
					}
				}(ccur, existing)
			}
		}
		wg.Wait()
	}

	go func() {
		cleanupped <- err
	}()
}

func (e *Executor) cleanupResource(cleanupper Cleanupper, resource string) error {
	if _, err := e.ensureGraph.GetVertex(resource); err == nil {
		return err
	}
	return cleanupper.Delete(resource)
}
