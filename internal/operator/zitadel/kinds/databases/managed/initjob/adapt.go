package initjob

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/job"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func AdaptFunc(
	namespace string,
	name string,
	image string,
	labels map[string]string,
	serviceAccountName string,
) (
	resources.QueryFunc,
	resources.DestroyFunc,
	error,
) {
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

	query, err := job.AdaptFuncToEnsure(jobDef)
	if err != nil {
		return nil, nil, err
	}

	destroy, err := job.AdaptFuncToDestroy(name, namespace)
	if err != nil {
		return nil, nil, err
	}

	return query, destroy, nil
}
