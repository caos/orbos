package orbiter

import (
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/tree"
)

type AdaptFunc func(monitor mntr.Monitor, finishedChan chan struct{}, desired *tree.Tree, current *tree.Tree) (QueryFunc, DestroyFunc, ConfigureFunc, bool, error)

type retAdapt struct {
	query     QueryFunc
	destroy   DestroyFunc
	configure ConfigureFunc
	migrate   bool
	err       error
}

func AdaptFuncGoroutine(adapt func() (QueryFunc, DestroyFunc, ConfigureFunc, bool, error)) (QueryFunc, DestroyFunc, ConfigureFunc, bool, error) {
	retChan := make(chan retAdapt)
	go func() {
		query, destroy, configure, migrate, err := adapt()
		retChan <- retAdapt{query, destroy, configure, migrate, err}
	}()
	ret := <-retChan
	return ret.query, ret.destroy, ret.configure, ret.migrate, ret.err
}
