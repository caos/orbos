package cmd

import (
	"fmt"

	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/resources"

	"github.com/caos/orbos/internal/operator/boom/api/v1beta2"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/toleration"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func Reconcile(monitor mntr.Monitor, k8sClient *kubernetes.Client, binaryVersion string, boomSpec *v1beta2.Boom) error {

	var (
		tolerations  toleration.Tolerations
		nodeselector map[string]string
		boomVersion  string
	)

	resources := resources.Resources(corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("250m"),
			corev1.ResourceMemory: resource.MustParse("256Mi"),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("128Mi"),
		},
	})
	if boomSpec != nil {
		boomVersion = boomSpec.Version
		tolerations = boomSpec.Tolerations
		nodeselector = boomSpec.NodeSelector
		if boomSpec.Resources != nil {
			resources = *boomSpec.Resources
		}
	}
	if boomVersion == "" {
		boomVersion = binaryVersion
		monitor.Info(fmt.Sprintf("No version set in boom.yml, so default version %s will get applied", binaryVersion))
	}

	recMonitor := monitor.WithField("version", boomVersion)

	if k8sClient.Available() {
		if err := kubernetes.EnsureBoomArtifacts(monitor, k8sClient, boomVersion, tolerations, nodeselector, &resources); err != nil {
			recMonitor.Error(errors.Wrap(err, "Failed to deploy boom into k8s-cluster"))
			return err
		}
		recMonitor.Info("Applied boom")
	} else {
		recMonitor.Info("Failed to connect to k8s")
	}

	return nil
}
