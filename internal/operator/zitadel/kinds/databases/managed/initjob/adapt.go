package initjob

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/job"
	"github.com/caos/orbos/internal/operator/zitadel"
	"github.com/caos/orbos/mntr"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func AdaptFunc(
	monitor mntr.Monitor,
	namespace string,
	name string,
	image string,
	labels map[string]string,
	serviceAccountName string,
	checkDBRunning zitadel.EnsureFunc,
) (
	zitadel.QueryFunc,
	zitadel.DestroyFunc,
	error,
) {
	internalMonitor := monitor.WithField("component", "initjob")

	certPath := "/cockroach/cockroach-certs"
	defaultMode := int32(256)

	internalLabels := make(map[string]string, 0)
	for k, v := range labels {
		internalLabels[k] = v
	}
	internalLabels["app.kubernetes.io/component"] = "iam-database-init"

	jobDef := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    internalLabels,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					ServiceAccountName: serviceAccountName,
					Containers: []corev1.Container{{
						Name:            name,
						Image:           image,
						ImagePullPolicy: "IfNotPresent",
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "client-certs",
							MountPath: certPath,
						}},
						Command: []string{
							"/cockroach/cockroach",
							"init",
							"--certs-dir=" + certPath,
							"--host=cockroachdb-0.cockroachdb",
						},
					}},
					RestartPolicy: "OnFailure",
					Volumes: []corev1.Volume{{
						Name: "client-certs",
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName:  "cockroachdb.client.root",
								DefaultMode: &defaultMode,
							},
						},
					}},
				},
			},
		},
	}

	destroy, err := job.AdaptFuncToDestroy(name, namespace)
	if err != nil {
		return nil, nil, err
	}

	destroyers := []zitadel.DestroyFunc{
		zitadel.ResourceDestroyToZitadelDestroy(destroy),
	}

	query, err := job.AdaptFuncToEnsure(jobDef)
	if err != nil {
		return nil, nil, err
	}

	queriers := []zitadel.QueryFunc{
		zitadel.EnsureFuncToQueryFunc(checkDBRunning),
		zitadel.ResourceQueryToZitadelQuery(query),
	}

	return func(k8sClient *kubernetes.Client, queried map[string]interface{}) (zitadel.EnsureFunc, error) {
			return zitadel.QueriersToEnsureFunc(internalMonitor, false, queriers, k8sClient, queried)
		},
		zitadel.DestroyersToDestroyFunc(internalMonitor, destroyers),
		nil
}
