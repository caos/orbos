package backup

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/cronjob"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/job"
	"github.com/caos/orbos/internal/operator/zitadel"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

func AdaptFunc(
	monitor mntr.Monitor,
	name string,
	namespace string,
	labels map[string]string,
	databases []string,
	checkDBReady zitadel.EnsureFunc,
	bucket string,
	cron string,
	secretName string,
	secretKey string,
	timestamp string,
	features []string,
) (
	queryFunc zitadel.QueryFunc,
	destroyFunc zitadel.DestroyFunc,
	err error,
) {
	defaultMode := int32(256)
	certPath := "/cockroach/cockroach-certs"
	secretPath := "/secrets/sa.json"
	backupPath := "/cockroach"
	backupNameEnv := "BACKUP_NAME"
	cronjobName := "backup-" + name

	backupCommands := make([]string, 0)
	if timestamp != "" {
		backupCommands = append(backupCommands, "export "+backupNameEnv+"="+timestamp)
	} else {
		backupCommands = append(backupCommands, "export "+backupNameEnv+"=$(date +%Y-%m-%dT%H:%M:%SZ)")
	}
	for _, database := range databases {
		backupCommands = append(backupCommands,
			strings.Join([]string{
				"/scripts/backup.sh",
				name,
				bucket,
				database,
				backupPath,
				secretPath,
				certPath,
				"${" + backupNameEnv + "}",
			}, " "))
	}

	jobSpecDef := batchv1.JobSpec{
		Template: corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				RestartPolicy: corev1.RestartPolicyNever,
				Containers: []corev1.Container{{
					Name:  name,
					Image: "docker.pkg.github.com/caos/orbos/crbackup:zitadel",
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
					Name: "public-github-packages",
				}},
			},
		},
	}

	destroyers := []zitadel.DestroyFunc{}
	queriers := []zitadel.QueryFunc{}

	cronJobDef := &v1beta1.CronJob{
		ObjectMeta: v1.ObjectMeta{
			Name:      cronjobName,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: v1beta1.CronJobSpec{
			Schedule:          cron,
			ConcurrencyPolicy: v1beta1.ForbidConcurrent,
			JobTemplate: v1beta1.JobTemplateSpec{
				Spec: jobSpecDef,
			},
		},
	}

	destroyCJ, err := cronjob.AdaptFuncToDestroy(namespace, cronjobName)
	if err != nil {
		return nil, nil, err
	}

	queryCJ, err := cronjob.AdaptFuncToEnsure(cronJobDef)
	if err != nil {
		return nil, nil, err
	}

	jobDef := batchv1.Job{
		ObjectMeta: v1.ObjectMeta{
			Name:      cronjobName,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: jobSpecDef,
	}

	destroyJ, err := job.AdaptFuncToDestroy(namespace, cronjobName)
	if err != nil {
		return nil, nil, err
	}

	queryJ, err := job.AdaptFuncToEnsure(&jobDef)
	if err != nil {
		return nil, nil, err
	}

	cleanupJ := func(k8sClient *kubernetes.Client) error {
		monitor.Info("waiting for backup to be completed")
		if err := k8sClient.WaitUntilJobCompleted(namespace, cronjobName, 60); err != nil {
			monitor.Error(errors.Wrap(err, "error while waiting for backup to be completed"))
			return err
		}
		monitor.Info("backup is completed, cleanup")
		if err := k8sClient.DeleteJob(namespace, cronjobName); err != nil {
			monitor.Error(errors.Wrap(err, "error while trying to cleanup backup"))
			return err
		}
		monitor.Info("restore backup is completed")
		return nil
	}

	for _, feature := range features {
		switch feature {
		case "backup":
			destroyers = append(destroyers,
				zitadel.ResourceDestroyToZitadelDestroy(destroyCJ),
			)
			queriers = append(queriers,
				zitadel.EnsureFuncToQueryFunc(checkDBReady),
				zitadel.ResourceQueryToZitadelQuery(queryCJ),
			)
		case "instantbackup":
			destroyers = append(destroyers,
				zitadel.ResourceDestroyToZitadelDestroy(destroyJ),
			)
			queriers = append(queriers,
				zitadel.EnsureFuncToQueryFunc(checkDBReady),
				zitadel.ResourceQueryToZitadelQuery(queryJ),
				zitadel.EnsureFuncToQueryFunc(cleanupJ),
			)
		}
	}

	return func(k8sClient *kubernetes.Client, queried map[string]interface{}) (zitadel.EnsureFunc, error) {
			return zitadel.QueriersToEnsureFunc(monitor, false, queriers, k8sClient, queried)
		},
		zitadel.DestroyersToDestroyFunc(monitor, destroyers),
		nil
}
