package boom

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/caos/orbos/internal/api/boom"

	v1 "github.com/caos/orbos/internal/api/boom/v1"
	"github.com/caos/orbos/internal/operator/boom/app"
	gconfig "github.com/caos/orbos/internal/operator/boom/application/applications/grafana/config"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/kubernetes"
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

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (res ctrl.Result, err error) {
	internalMonitor := r.Monitor.WithFields(map[string]interface{}{
		"kind": "Boom",
		"name": req.NamespacedName,
	})

	defer func() {
		r.Monitor.Error(err)
	}()

	if req.Namespace != boom.Name || req.Name != boom.Name {
		return res, fmt.Errorf("resource must be named %s and namespaced in %s", boom.Name, boom.Namespace)
	}

	desired, err := boom.ReadCRD(r)
	if err != nil {
		return res, err
	}

	gconfig.DashboardsDirectoryPath = filepath.Join(r.ToolsDirectoryPath, "dashboards")
	appStruct := app.New(internalMonitor, r.ToolsDirectoryPath)

	return res, appStruct.ReconcileCrd(req.NamespacedName.String(), desired)
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.Boom{}).
		Complete(r)
}
