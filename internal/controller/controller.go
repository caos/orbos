package controller

import (
	"github.com/caos/orbos/internal/controller/database"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	Database   = "database"
	Networking = "networking"
)

func Start(monitor mntr.Monitor, metricsAddr string, features ...string) error {
	scheme := runtime.NewScheme()

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
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
				Client: kubernetes.NewK8sClient(),
				Log:    ctrl.Log.WithName("controllers").WithName("IAM"),
				Scheme: mgr.GetScheme(),
			}).SetupWithManager(mgr); err != nil {
				setupLog.Error(err, "unable to create controller", "controller", "IAM")
				os.Exit(1)
			}

		}
	}
}
