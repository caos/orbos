package resources

import (
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/tree"
	"github.com/pkg/errors"
)

type AdaptFuncToEnsure func(monitor mntr.Monitor, desired *tree.Tree, current *tree.Tree) (QueryFunc, error)
type AdaptFuncToDelete func(monitor mntr.Monitor, desired *tree.Tree, current *tree.Tree) (DestroyFunc, error)

type Querier interface {
	Query(k8sClient kubernetes.ClientInt, queried map[string]interface{}) (EnsureFunc, error)
}

type QueryFunc func(k8sClient kubernetes.ClientInt, queried map[string]interface{}) (EnsureFunc, error)

func (q QueryFunc) Query(k8sClient kubernetes.ClientInt, queried map[string]interface{}) (EnsureFunc, error) {
	return q(k8sClient, queried)
}

type Ensurer interface {
	Ensure(k8sClient kubernetes.ClientInt) error
}

type EnsureFunc func(k8sClient kubernetes.ClientInt) error

func (e EnsureFunc) Ensure(k8sClient kubernetes.ClientInt) error {
	return e(k8sClient)
}

type Destroyer interface {
	Destroy(k8sClient kubernetes.ClientInt) error
}

type DestroyFunc func(k8sClient kubernetes.ClientInt) error

func (d DestroyFunc) Destroy(k8sClient kubernetes.ClientInt) error {
	return d(k8sClient)
}

type TestableDestroyer struct {
	DestroyFunc
	Arguments interface{}
}

func ToTestableDestroy(destroyFunc DestroyFunc, adapterArguments interface{}) *TestableDestroyer {
	return &TestableDestroyer{
		DestroyFunc: destroyFunc,
		Arguments:   adapterArguments,
	}
}

type TestableQuerier struct {
	QueryFunc
	Arguments interface{}
}

func ToTestableQuerier(queryFunc QueryFunc, adapterArguments interface{}) *TestableQuerier {
	return &TestableQuerier{
		QueryFunc: queryFunc,
		Arguments: adapterArguments,
	}
}

func WrapFuncs(monitor mntr.Monitor, query QueryFunc, destroy DestroyFunc) (QueryFunc, DestroyFunc, error) {
	return func(client kubernetes.ClientInt, queried map[string]interface{}) (ensureFunc EnsureFunc, err error) {
			monitor.Info("querying...")
			ensurer, err := query(client, queried)
			if err != nil {
				err := errors.Wrapf(err, "error while querying")
				monitor.Error(err)
				return nil, err
			}
			monitor.Info("queried")
			return func(k8sClient kubernetes.ClientInt) error {
				monitor.Info("ensuring...")
				if err := ensurer.Ensure(k8sClient); err != nil {
					return errors.Wrap(err, "error while destroying")
				}
				monitor.Info("ensured")
				return nil
			}, nil
		}, func(client kubernetes.ClientInt) error {
			monitor.Info("destroying...")
			err := destroy(client)
			if err != nil {
				err := errors.Wrapf(err, "error while destroying")
				monitor.Error(err)
				return err
			}
			monitor.Info("destroyed")
			return nil
		}, nil
}

var _ QueriersReducerFunc = QueriersToEnsurer

type QueriersReducerFunc func(monitor mntr.Monitor, infoLogs bool, queriers []Querier, k8sClient kubernetes.ClientInt, queried map[string]interface{}) (EnsureFunc, error)

func QueriersToEnsurer(monitor mntr.Monitor, infoLogs bool, queriers []Querier, k8sClient kubernetes.ClientInt, queried map[string]interface{}) (EnsureFunc, error) {
	if infoLogs {
		monitor.Info("querying...")
	} else {
		monitor.Debug("querying...")
	}
	ensurers := make([]EnsureFunc, 0)
	for _, querier := range queriers {
		ensurer, err := querier.Query(k8sClient, queried)
		if err != nil {
			return nil, errors.Wrap(err, "error while querying")
		}
		ensurers = append(ensurers, ensurer)
	}
	if infoLogs {
		monitor.Info("queried")
	} else {
		monitor.Debug("queried")
	}
	return func(k8sClient kubernetes.ClientInt) error {
		if infoLogs {
			monitor.Info("ensuring...")
		} else {
			monitor.Debug("ensuring...")
		}
		for _, ensurer := range ensurers {
			if err := ensurer.Ensure(k8sClient); err != nil {
				return errors.Wrap(err, "error while ensuring")
			}
		}
		if infoLogs {
			monitor.Info("ensured")
		} else {
			monitor.Debug("ensured")
		}
		return nil
	}, nil
}

func ReduceDestroyers(monitor mntr.Monitor, destroyers []Destroyer) DestroyFunc {
	return func(k8sClient kubernetes.ClientInt) error {
		monitor.Info("destroying...")
		for _, destroyer := range destroyers {
			if err := destroyer.Destroy(k8sClient); err != nil {
				return errors.Wrap(err, "error while destroying")
			}
		}
		monitor.Info("destroyed")
		return nil
	}
}

func DestroyerToQuerier(destroyer Destroyer) QueryFunc {
	return func(k8sClient kubernetes.ClientInt, queried map[string]interface{}) (ensureFunc EnsureFunc, err error) {
		return destroyer.Destroy, nil
	}
}

func DestroyersToQueryFuncs(destroyers []Destroyer) []QueryFunc {
	queriers := make([]QueryFunc, len(destroyers))
	for i, destroyer := range destroyers {
		queriers[i] = DestroyerToQuerier(destroyer)
	}
	return queriers
}
