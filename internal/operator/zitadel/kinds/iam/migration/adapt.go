package migration

import (
	"crypto/sha1"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/configmap"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/job"
	"github.com/caos/orbos/internal/operator/zitadel"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/iam/migration/scripts"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

func AdaptFunc(
	namespace string,
	labels map[string]string,
	secretPasswordName string,
	migrationUser string,
	users []string,
) (
	zitadel.QueryFunc,
	zitadel.DestroyFunc,
	error,
) {
	migrationConfigmap := "migrate-db"
	migrationsPath := "/migrate"
	migrationsConfigInternal := "migrate-db"
	rootUserInternal := "root"
	rootUserPath := "/certificates"
	defaultMode := int32(0400)

	internalLabels := make(map[string]string, 0)
	for k, v := range labels {
		internalLabels[k] = v
	}
	internalLabels["app.kubernetes.io/component"] = "migration"
	allScripts := scripts.GetAll()
	scriptsStr := ""
	for k, v := range allScripts {
		if scriptsStr == "" {
			scriptsStr = k + ": " + v
		} else {
			scriptsStr = scriptsStr + "," + k + ": " + v
		}
	}
	h := sha1.New()
	_, err := h.Write([]byte(scriptsStr))
	if err != nil {
		return nil, nil, err
	}
	hash := h.Sum(nil)

	queryCM, destroyCM, err := configmap.AdaptFunc(migrationConfigmap, namespace, labels, allScripts)
	if err != nil {
		return nil, nil, err
	}

	envMigrationUser := "FLYWAY_USER"
	envMigrationPW := "FLYWAY_PASSWORD"

	envVars := []corev1.EnvVar{
		{
			Name:  envMigrationUser,
			Value: migrationUser,
		}, {
			Name: envMigrationPW,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: secretPasswordName},
					Key:                  migrationUser,
				},
			},
		},
	}

	migrationEnvVars := make([]corev1.EnvVar, 0)
	for _, v := range envVars {
		migrationEnvVars = append(migrationEnvVars, v)
	}
	for _, user := range users {
		migrationEnvVars = append(migrationEnvVars, corev1.EnvVar{
			Name: "FLYWAY_PLACEHOLDERS_" + strings.ToUpper(user) + "PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: secretPasswordName},
					Key:                  migrationUser,
				},
			},
		})
	}

	createFile := "create.sql"
	createStr := strings.Join([]string{
		"echo -n 'CREATE USER IF NOT EXISTS ' > " + createFile,
		"echo -n ${" + envMigrationUser + "} >> " + createFile,
		"echo -n ';' >> " + createFile,
		"echo -n 'ALTER USER ' >> " + createFile,
		"echo -n ${" + envMigrationUser + "} >> " + createFile,
		"echo -n ' WITH PASSWORD ' >> " + createFile,
		"echo -n ${" + envMigrationPW + "} >> " + createFile,
		"echo -n ';' >> " + createFile,
	}, ";")
	grantFile := "grant.sql"
	grantStr := strings.Join([]string{
		"echo -n 'GRANT admin TO ' > " + grantFile,
		"echo -n ${" + envMigrationUser + "} >> " + grantFile,
		"echo -n ' WITH ADMIN OPTION;'  >> " + grantFile,
	}, ";")
	deleteFile := "delete.sql"
	deleteStr := strings.Join([]string{
		"echo -n 'DROP USER IF EXISTS ' > " + deleteFile,
		"echo -n ${" + envMigrationUser + "} >> " + deleteFile,
		"echo -n ';' >> " + deleteFile,
	}, ";")

	jobDef := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cockroachdb-cluster-migration",
			Namespace: namespace,
			Labels:    internalLabels,
			Annotations: map[string]string{
				"migrationhash": string(hash),
			},
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
								"until pg_isready -h cockroachdb-public -p 26257; do echo waiting for database; sleep 2; done;",
							},
						},
						{
							Name:  "create-flyway-user",
							Image: "cockroachdb/cockroach:v20.1.3",
							Env:   envVars,
							VolumeMounts: []corev1.VolumeMount{{
								Name:      rootUserInternal,
								MountPath: rootUserPath,
							}},
							Command: []string{"/bin/bash", "-c", "--"},
							Args: []string{
								strings.Join([]string{
									createStr,
									grantStr,
									"cockroach.sh sql --certs-dir=/certificates --host=cockroachdb-public:26257 -e \"$(cat " + createFile + ")\" -e \"$(cat " + grantFile + ")\";",
								},
									";"),
							},
						},
						{
							Name:            "db-migration",
							Image:           "flyway/flyway:6.5.0",
							ImagePullPolicy: "Always",
							Args: []string{
								"-url=jdbc:postgresql://cockroachdb-public:26257/defaultdb?&sslmode=verify-full&ssl=true&sslrootcert=" + rootUserPath + "/ca.crt&sslfactory=org.postgresql.ssl.NonValidatingFactory",
								"-locations=filesystem:" + migrationsPath,
								"migrate",
							},
							Env: migrationEnvVars,
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
							Name:    "delete-flyway-user",
							Image:   "cockroachdb/cockroach:v20.1.3",
							Command: []string{"/bin/bash", "-c", "--"},
							Args: []string{
								strings.Join([]string{
									deleteStr,
									"cockroach.sh sql --certs-dir=/certificates --host=cockroachdb-public:26257 -e \"$(cat " + deleteFile + ")\";",
								}, ";"),
							},
							Env: envVars,
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
					}, {
						Name: secretPasswordName,
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: secretPasswordName,
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

	queriers := []zitadel.QueryFunc{
		zitadel.ResourceQueryToZitadelQuery(queryCM),
		zitadel.ResourceQueryToZitadelQuery(queryJ),
	}
	destroyers := []zitadel.DestroyFunc{
		zitadel.ResourceDestroyToZitadelDestroy(destroyJ),
		zitadel.ResourceDestroyToZitadelDestroy(destroyCM),
	}

	return func(k8sClient *kubernetes.Client, queried map[string]interface{}) (zitadel.EnsureFunc, error) {
			return zitadel.QueriersToEnsureFunc(queriers, k8sClient, queried)
		},
		zitadel.DestroyersToDestroyFunc(destroyers),
		nil
}

//sql --certs-dir = /certificates --host =cockroachdb-public:26257 -e CREATE USER IF NOT EXISTS ${FLYWAY_USER} WITH PASSWORD ${FLYWAY_PASSWORD}; -e GRANT admin TO ${FLYWAY_USER} WITH ADMIN OPTION
