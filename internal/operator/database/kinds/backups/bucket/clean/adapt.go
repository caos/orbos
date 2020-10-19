package clean

import (
	"github.com/caos/orbos/internal/operator/core"
	"github.com/caos/orbos/mntr"
	kubernetes2 "github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/kubernetes/resources/job"
	"github.com/pkg/errors"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

func ApplyFunc(
	monitor mntr.Monitor,
	name string,
	namespace string,
	labels map[string]string,
	databases []string,
	nodeselector map[string]string,
	tolerations []corev1.Toleration,
	checkDBReady core.EnsureFunc,
	secretName string,
	secretKey string,
	version string,
	imagePullSecretName string,
) (
	queryFunc core.QueryFunc,
	destroyFunc core.DestroyFunc,
	ensureFunc core.EnsureFunc,
	err error,
) {
	defaultMode := int32(256)
	certPath := "/cockroach/cockroach-certs"
	secretPath := "/secrets/sa.json"

	jobName := "backup-" + name + "-clean"

	backupCommands := make([]string, 0)
	for _, database := range databases {
		backupCommands = append(backupCommands,
			strings.Join([]string{
				"/scripts/clean-db.sh",
				certPath,
				database,
			}, " "))
	}
	for _, database := range databases {
		backupCommands = append(backupCommands,
			strings.Join([]string{
				"/scripts/clean-user.sh",
				certPath,
				database,
			}, " "))
	}
	backupCommands = append(backupCommands,
		strings.Join([]string{
			"/scripts/clean-migration.sh",
			certPath,
		}, " "))

	jobdef := &batchv1.Job{
		ObjectMeta: v1.ObjectMeta{
			Name:      jobName,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					NodeSelector:  nodeselector,
					Tolerations:   tolerations,
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{{
						Name:  jobName,
						Image: "docker.pkg.github.com/caos/orbos/crbackup:" + version,
						Command: []string{
							"/bin/bash",
							"-c",
							strings.Join(backupCommands, " && "),
						},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "client-certs",
							MountPath: certPath,
						}, {
							Name:      secretKey,
							SubPath:   secretKey,
							MountPath: secretPath,
						}},
						ImagePullPolicy: corev1.PullAlways,
					}},
					Volumes: []corev1.Volume{{
						Name: "client-certs",
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName:  "cockroachdb.client.root",
								DefaultMode: &defaultMode,
							},
						},
					}, {
						Name: secretKey,
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: secretName,
							},
						},
					}},
					ImagePullSecrets: []corev1.LocalObjectReference{{
						Name: imagePullSecretName,
					}},
				},
			},
		},
	}

	destroyJ, err := job.AdaptFuncToDestroy(jobName, namespace)
	if err != nil {
		return nil, nil, nil, err
	}

	destroyers := []core.DestroyFunc{
		core.ResourceDestroyToZitadelDestroy(destroyJ),
	}

	queryJ, err := job.AdaptFuncToEnsure(jobdef)
	if err != nil {
		return nil, nil, nil, err
	}

	queriers := []core.QueryFunc{
		core.EnsureFuncToQueryFunc(checkDBReady),
		core.ResourceQueryToZitadelQuery(queryJ),
	}

	return func(k8sClient *kubernetes2.Client, queried map[string]interface{}) (core.EnsureFunc, error) {
			return core.QueriersToEnsureFunc(monitor, false, queriers, k8sClient, queried)
		},
		core.DestroyersToDestroyFunc(monitor, destroyers),
		func(k8sClient *kubernetes2.Client) error {
			monitor.Info("waiting for clean to be completed")
			if err := k8sClient.WaitUntilJobCompleted(namespace, jobName, 60); err != nil {
				monitor.Error(errors.Wrap(err, "error while waiting for clean to be completed"))
				return err
			}
			monitor.Info("clean is completed, cleanup")
			if err := k8sClient.DeleteJob(namespace, jobName); err != nil {
				monitor.Error(errors.Wrap(err, "error while trying to cleanup clean"))
				return err
			}
			monitor.Info("clean cleanup is completed")
			return nil
		},
		nil
}
