package orbiter

import (
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/secret"
	"github.com/caos/orbos/pkg/tree"
)

type AdaptFunc func(monitor mntr.Monitor, finishedChan chan struct{}, desired *tree.Tree, current *tree.Tree) (QueryFunc, DestroyFunc, ConfigureFunc, bool, map[string]*secret.Secret, error)

type retAdapt struct {
	query     QueryFunc
	destroy   DestroyFunc
	configure ConfigureFunc
	migrate   bool
	secrets   map[string]*secret.Secret
	err       error
}

func AdaptFuncGoroutine(adapt func() (QueryFunc, DestroyFunc, ConfigureFunc, bool, map[string]*secret.Secret, error)) (QueryFunc, DestroyFunc, ConfigureFunc, bool, map[string]*secret.Secret, error) {
	retChan := make(chan retAdapt)
	go func() {
		query, destroy, configure, migrate, secret, err := adapt()
		retChan <- retAdapt{query, destroy, configure, migrate, secret, err}
	}()
	ret := <-retChan
	return ret.query, ret.destroy, ret.configure, ret.migrate, ret.secrets, ret.err
}
