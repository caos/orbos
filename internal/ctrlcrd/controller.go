package ctrlcrd

import (
	"fmt"

	boomv1 "github.com/caos/orbos/internal/api/boom/v1"
	networkingv1 "github.com/caos/orbos/internal/api/networking/v1"
	"github.com/caos/orbos/internal/ctrlcrd/boom"
	"github.com/caos/orbos/internal/ctrlcrd/networking"
	"github.com/caos/orbos/internal/utils/clientgo"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/kubernetes"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	Networking = "networking"
	Boom       = "boom"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		panic(fmt.Errorf("adding clientgo to scheme failed: %w", err))
	}

	if err := networkingv1.AddToScheme(scheme); err != nil {
		panic(fmt.Errorf("adding networking v1 to scheme failed: %w", err))
	}

	if err := boomv1.AddToScheme(scheme); err != nil {
		panic(fmt.Errorf("adding boom v1 to scheme failed: %w", err))
	}

}

func Start(monitor mntr.Monitor, version, toolsDirectoryPath, metricsAddr string, kubeconfig string, features ...string) error {

	cfg, err := clientgo.GetClusterConfig(monitor, kubeconfig)
	if err != nil {
		return err
	}

	k8sClient, err := kubernetes.NewK8sClientWithConfig(monitor, cfg)
	if err != nil {
		return err
	}

	monitor.Info("successfully connected to kubernetes cluster")

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     false,
		LeaderElectionID:   "98jasd12l.caos.ch",
	})
	if err != nil {
		return fmt.Errorf("starting manager failed: %w", err)
	}

	for _, feature := range features {
		switch feature {
		case Networking:
			monitor.Debug("Setting up networking")
			if err = (&networking.Reconciler{
				ClientInt: k8sClient,
				Monitor:   monitor,
				Scheme:    mgr.GetScheme(),
				Version:   version,
			}).SetupWithManager(mgr); err != nil {
				return fmt.Errorf("creating controller failed: %w", err)
			}
			monitor.Debug("Networking setup done")
		case Boom:
			monitor.Debug("Setting up BOOM")
			if err = (&boom.Reconciler{
				ClientInt:          k8sClient,
				Monitor:            monitor,
				Scheme:             mgr.GetScheme(),
				ToolsDirectoryPath: toolsDirectoryPath,
				Version:            version,
			}).SetupWithManager(mgr); err != nil {
				return fmt.Errorf("creating controller failed: %w", err)
			}
			monitor.Debug("BOOM setup done")
		}
	}

	monitor.Debug("Controller is starting")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		return fmt.Errorf("running manager failed: %w", err)
	}
	monitor.Debug("Controller is done")
	return nil
}
