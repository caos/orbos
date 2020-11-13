package cmd

import (
	"github.com/caos/orbos/internal/operator/boom/api/latest"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/kubernetes/k8s"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func Reconcile(monitor mntr.Monitor, k8sClient *kubernetes.Client, boomSpec *latest.Boom) error {

	resources := k8s.Resources(corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("250m"),
			corev1.ResourceMemory: resource.MustParse("512Mi"),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("256Mi"),
		},
	})

	var imageRegistry string
	var nodeselector map[string]string
	var tolerations k8s.Tolerations
	if boomSpec != nil {
		if boomSpec.Version == "" {
			err := errors.New("No version set in boom.yml")
			monitor.Error(err)
			return err
		}

		if boomSpec.Resources != nil {
			resources = *boomSpec.Resources
		}
		if boomSpec.NodeSelector != nil {
			nodeselector = boomSpec.NodeSelector
		}
		if boomSpec.Tolerations != nil {
			tolerations = boomSpec.Tolerations
		}
		imageRegistry = boomSpec.CustomImageRegistry
	}
	recMonitor := monitor.WithField("version", boomSpec.Version)
	if imageRegistry == "" {
		imageRegistry = "ghcr.io"
	}

	if !k8sClient.Available() {
		recMonitor.Info("Failed to connect to k8s")
		return nil
	}

	if err := kubernetes.EnsureBoomArtifacts(monitor, k8sClient, boomSpec.Version, tolerations, nodeselector, &resources, imageRegistry); err != nil {
		recMonitor.Error(errors.Wrap(err, "Failed to deploy boom into k8s-cluster"))
		return err
	}
	return nil
}
