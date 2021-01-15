package boom

import (
	"context"
	"errors"
	v1 "github.com/caos/orbos/internal/api/networking/v1"
	orbnw "github.com/caos/orbos/internal/operator/networking/kinds/orb"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/tree"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Reconciler struct {
	kubernetes.ClientInt
	Monitor mntr.Monitor
	Scheme  *runtime.Scheme
	Version string
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	internalMonitor := r.Monitor.WithFields(map[string]interface{}{
		"kind":      "boom",
		"namespace": req.NamespacedName,
	})

	unstruct, err := r.ClientInt.GetNamespacedCRDResource(v1.GroupVersion.Group, v1.GroupVersion.Version, "Boom", req.Namespace, req.Name)
	if err != nil {
		return ctrl.Result{}, err
	}

	spec, found := unstruct.Object["spec"]
	if !found {
		return ctrl.Result{}, errors.New("no spec in crd")
	}
	specMap, ok := spec.(map[string]interface{})
	if !ok {
		return ctrl.Result{}, errors.New("no spec in crd")
	}

	//TODO

	return ctrl.Result{}, nil
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.Networking{}).
		Complete(r)
}
