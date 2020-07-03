package migration

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/configmap"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/job"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/iam/migration/scripts"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func AdaptFunc(
	namespace string,
	labels map[string]string,
	migrationConfigmap string,
) (
	resources.QueryFunc,
	resources.DestroyFunc,
	error,
) {

	queriers := make([]resources.QueryFunc, 0)
	destroyers := make([]resources.DestroyFunc, 0)

	queryCM, destroyCM, err := configmap.AdaptFunc(migrationConfigmap, namespace, labels, scripts.GetAll())
	if err != nil {
		return nil, nil, err
	}

	migrationsPath := "/migrate"
	migrationsConfigInternal := "migrate-db"
	rootUserInternal := "root"
	rootUserPath := "/certificates"
	defaultMode := int32(0400)

	jobDef := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cockroachdb-cluster-migration",
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						{
							Name:  "check-db-ready",
							Image: "postgres:9.6.17",
							Command: []string{
								"sh",
								"-c",
								"sleep 10; until pg_isready -h cockroachdb-public -p 26257; do echo waiting for database; sleep 2; done; sleep 10;",
							},
						},
						{
							Name:  "create-flyway-user",
							Image: "cockroachdb/cockroach:v20.1.2",
							Command: []string{
								"sh",
								"-c",
								"cockroach sql --certs-dir=" + rootUserPath + " --host=cockroachdb-public:26257 -e \"CREATE USER IF NOT EXISTS flyway WITH PASSWORD flyway;\" -e \"GRANT admin TO flyway;\"",
							},
							VolumeMounts: []corev1.VolumeMount{{
								Name:      rootUserInternal,
								MountPath: rootUserPath,
							}},
						},
						{
							Name:            "db-migration",
							Image:           "flyway/flyway:6.5.0",
							ImagePullPolicy: "Always",
							Args: []string{
								"-url=jdbc:postgresql://cockroachdb-public:26257/defaultdb?&user=flyway&sslmode=verify-full&ssl=true&sslrootcert=" + rootUserPath + "/ca.crt&sslfactory=org.postgresql.ssl.NonValidatingFactory&sslcert=" + rootUserPath + "/client.root.crt&sslkey=" + rootUserPath + "/client.root.key",
								"-locations=filesystem:" + migrationsPath,
								"-user=flyway",
								"-password=flyway",
								"migrate",
							},

							VolumeMounts: []corev1.VolumeMount{{
								Name:      migrationsConfigInternal,
								MountPath: migrationsPath,
							}, {
								Name:      rootUserInternal,
								MountPath: rootUserPath,
							}},
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "delete-flyway-user",
							Image: "cockroachdb/cockroach:v20.1.2",
							Command: []string{
								"sh",
								"-c",
								"cockroach sql --certs-dir=" + rootUserPath + " --host=cockroachdb-public:26257 -e \"DROP USER IF EXISTS flyway;\"",
							},
							VolumeMounts: []corev1.VolumeMount{{
								Name:      rootUserInternal,
								MountPath: rootUserPath,
							}},
						},
					},
					RestartPolicy: "Never",
					Volumes: []corev1.Volume{{
						Name: migrationsConfigInternal,
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{Name: migrationConfigmap},
							},
						},
					}, {
						Name: rootUserInternal,
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

	queryJ, destroyJ, err := job.AdaptFunc(jobDef)
	if err != nil {
		return nil, nil, err
	}

	queriers = append(queriers, queryCM, queryJ)
	destroyers = append(destroyers, destroyCM, destroyJ)

	return func() (resources.EnsureFunc, error) {
			ensurers := make([]resources.EnsureFunc, 0)
			for _, querier := range queriers {
				ensurer, err := querier()
				if err != nil {
					return nil, err
				}
				ensurers = append(ensurers, ensurer)
			}

			return func(k8sClient *kubernetes.Client) error {
				for _, ensurer := range ensurers {
					if err := ensurer(k8sClient); err != nil {
						return err
					}
				}
				return nil
			}, nil
		}, func(k8sClient *kubernetes.Client) error {
			for _, destroyer := range destroyers {
				if err := destroyer(k8sClient); err != nil {
					return err
				}
			}
			return nil
		},
		nil
}
