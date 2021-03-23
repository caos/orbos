package cmd

import (
	"fmt"

	"github.com/caos/orbos/internal/operator/boom/api/latest"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/kubernetes/k8s"
	"github.com/caos/orbos/pkg/labels"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func Reconcile(
	monitor mntr.Monitor,
	apiLabels *labels.API,
	k8sClient kubernetes.ClientInt,
	boomSpec *latest.Boom,
	binaryVersion string,
	gitops bool,
) error {

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

	imageRegistry := "ghcr.io"
	var boomVersion string
	var nodeselector map[string]string
	var tolerations k8s.Tolerations

	if boomSpec != nil {
		if boomSpec.Version != "" {
			boomVersion = boomSpec.Version
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
		if boomSpec.CustomImageRegistry != "" {
			imageRegistry = boomSpec.CustomImageRegistry
		}
	}
	if boomVersion == "" {
		monitor.Info(fmt.Sprintf("No version set in boom.yml, so current binary version %s will get applied", binaryVersion))
		boomVersion = binaryVersion
	}

	recMonitor := monitor.WithField("version", boomVersion)

	if err := kubernetes.EnsureBoomArtifacts(monitor, apiLabels, k8sClient, boomVersion, tolerations, nodeselector, &resources, imageRegistry, gitops); err != nil {
		recMonitor.Error(errors.Wrap(err, "Failed to deploy boom into k8s-cluster"))
		return err
	}

	recMonitor.Info("Applied BOOM")

	return nil
}
