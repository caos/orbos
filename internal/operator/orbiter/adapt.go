package orbiter

import (
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
)

type AdaptFunc func(monitor mntr.Monitor, finishedChan chan bool, desired *tree.Tree, current *tree.Tree) (QueryFunc, DestroyFunc, bool, error)

type retAdapt struct {
	query   QueryFunc
	destroy DestroyFunc
	migrate bool
	err     error
}

func AdaptFuncGoroutine(adapt func() (QueryFunc, DestroyFunc, bool, error)) (QueryFunc, DestroyFunc, bool, error) {
	retChan := make(chan retAdapt)
	go func() {
		query, destroy, migrate, err := adapt()
		retChan <- retAdapt{query, destroy, migrate, err}
	}()
	ret := <-retChan
	return ret.query, ret.destroy, ret.migrate, ret.err
}
