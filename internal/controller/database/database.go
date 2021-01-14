package database

import (
	"context"
	v1 "github.com/caos/orbos/internal/api/database/v1"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

func Start() {

}

type Reconciler struct {
	kubernetes.ClientInt
	monitor mntr.Monitor
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	internalMonitor := r.monitor.WithFields(map[string]interface{}{
		"kind":      "iam",
		"namespace": req.NamespacedName,
	})

	// TODO

	return ctrl.Result{}, nil
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.Database{}).
		Complete(r)
}
