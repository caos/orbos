package controller

import (
	boomv1 "github.com/caos/orbos/internal/api/boom/v1"
	databasev1 "github.com/caos/orbos/internal/api/database/v1"
	networkingv1 "github.com/caos/orbos/internal/api/networking/v1"
	"github.com/caos/orbos/internal/controller/boom"
	"github.com/caos/orbos/internal/controller/database"
	"github.com/caos/orbos/internal/controller/networking"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	Database   = "database"
	Networking = "networking"
	Boom       = "boom"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = databasev1.AddToScheme(scheme)
	_ = networkingv1.AddToScheme(scheme)
	_ = boomv1.AddToScheme(scheme)
}

func Start(monitor mntr.Monitor, version, toolsDirectoryPath, metricsAddr string, features ...string) error {
	cfg := ctrl.GetConfigOrDie()
	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     false,
		LeaderElectionID:   "98jasd12l.caos.ch",
	})
	if err != nil {
		return errors.Wrap(err, "unable to start manager")
	}

	for _, feature := range features {
		switch feature {
		case Database:
			if err = (&database.Reconciler{
				ClientInt: kubernetes.NewK8sClientWithConfig(monitor, cfg),
				Monitor:   monitor,
				Scheme:    mgr.GetScheme(),
				Version:   version,
			}).SetupWithManager(mgr); err != nil {
				return errors.Wrap(err, "unable to create controller")
			}
		case Networking:
			if err = (&networking.Reconciler{
				ClientInt: kubernetes.NewK8sClientWithConfig(monitor, cfg),
				Monitor:   monitor,
				Scheme:    mgr.GetScheme(),
				Version:   version,
			}).SetupWithManager(mgr); err != nil {
				return errors.Wrap(err, "unable to create controller")
			}
		case Boom:
			if err = (&boom.Reconciler{
				ClientInt:          kubernetes.NewK8sClientWithConfig(monitor, cfg),
				Monitor:            monitor,
				Scheme:             mgr.GetScheme(),
				ToolsDirectoryPath: toolsDirectoryPath,
				Version:            version,
			}).SetupWithManager(mgr); err != nil {
				return errors.Wrap(err, "unable to create controller")
			}

		}
	}

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		return errors.Wrap(err, "problem running manager")
	}
	return nil
}
