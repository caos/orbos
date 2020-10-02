package cmd

import (
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/kubernetes/k8s"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func Reconcile(monitor mntr.Monitor, k8sClient *kubernetes.Client, boomSpec *v1beta2.Boom) error {
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

	if boomSpec != nil {
		if boomSpec.Version == "" {
			err := errors.New("No version set in boom.yml")
			monitor.Error(err)
			return err
		}

		if boomSpec.Resources != nil {
			resources = *boomSpec.Resources
		}

		recMonitor := monitor.WithField("version", boomSpec.Version)

		if k8sClient.Available() {
			if err := kubernetes.EnsureBoomArtifacts(monitor, k8sClient, boomSpec.Version, boomSpec.Tolerations, boomSpec.NodeSelector, &resources); err != nil {
				recMonitor.Error(errors.Wrap(err, "Failed to deploy boom into k8s-cluster"))
				return err
			}
			recMonitor.Info("Applied boom")
		} else {
			recMonitor.Info("Failed to connect to k8s")
		}
	}
	return nil
}
