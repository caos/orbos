package boom

import (
	"context"
	"path/filepath"

	v1 "github.com/caos/orbos/internal/api/boom/v1"
	"github.com/caos/orbos/internal/operator/boom/app"
	gconfig "github.com/caos/orbos/internal/operator/boom/application/applications/grafana/config"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/tree"
	"gopkg.in/yaml.v3"
	macherrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Reconciler struct {
	kubernetes.ClientInt
	Monitor            mntr.Monitor
	Scheme             *runtime.Scheme
	ToolsDirectoryPath string
	Version            string
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	internalMonitor := r.Monitor.WithFields(map[string]interface{}{
		"kind": "BOOM",
		"name": req.NamespacedName,
	})

	unstruct, err := r.ClientInt.GetNamespacedCRDResource(v1.GroupVersion.Group, v1.GroupVersion.Version, "BOOM", req.Namespace, req.Name)
	if !macherrs.IsNotFound(err) && err != nil {
		return ctrl.Result{}, err
	}

	var data []byte
	if macherrs.IsNotFound(err) {
		unstruct = &unstructured.Unstructured{Object: v1.GetEmpty(req.Namespace, req.Name)}
		err = nil
	}

	dataInt, err := yaml.Marshal(unstruct.Object)
	if err != nil {
		return ctrl.Result{}, err
	}
	data = dataInt

	desired := &tree.Tree{}
	if err := yaml.Unmarshal(data, &desired); err != nil {
		return ctrl.Result{}, err
	}

	gconfig.DashboardsDirectoryPath = filepath.Join(r.ToolsDirectoryPath, "dashboards")
	appStruct := app.New(internalMonitor, r.ToolsDirectoryPath)

	if err := appStruct.ReconcileCrd(req.NamespacedName.String(), desired); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.BOOM{}).
		Complete(r)
}
