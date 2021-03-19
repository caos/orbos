package networking

import (
	"context"

	"github.com/caos/orbos/internal/api/networking"
	v1 "github.com/caos/orbos/internal/api/networking/v1"
	orbnw "github.com/caos/orbos/internal/operator/networking/kinds/orb"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/tree"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Reconciler struct {
	kubernetes.ClientInt
	Monitor mntr.Monitor
	Scheme  *runtime.Scheme
	Version string
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (res ctrl.Result, err error) {
	internalMonitor := r.Monitor.WithFields(map[string]interface{}{
		"kind": "networking",
	})

	desired, err := networking.ReadCRD(r)
	if desired == nil {
		return res, err
	}

	query, _, _, _, err := orbnw.AdaptFunc(&r.Version, false)(internalMonitor, desired, &tree.Tree{})
	if err != nil {
		return res, err
	}

	ensure, err := query(r.ClientInt, map[string]interface{}{})
	if err != nil {
		return res, err
	}

	return res, ensure(r.ClientInt)
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.Networking{}).
		Complete(r)
}
