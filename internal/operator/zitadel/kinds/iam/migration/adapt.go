package migration

import (
	"crypto/sha1"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/configmap"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/job"
	"github.com/caos/orbos/internal/operator/zitadel"
	coredb "github.com/caos/orbos/internal/operator/zitadel/kinds/databases/core"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/iam/migration/scripts"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

func AdaptFunc(
	monitor mntr.Monitor,
	namespace string,
	labels map[string]string,
	secretPasswordName string,
	migrationUser string,
	users []string,
) (
	zitadel.QueryFunc,
	zitadel.DestroyFunc,
	zitadel.EnsureFunc,
	error,
) {
	internalMonitor := monitor.WithField("component", "migration")

	migrationConfigmap := "migrate-db"
	migrationsPath := "/migrate"
	rootUserInternal := "root"
	rootUserPath := "/certificates"
	defaultMode := int32(0400)
	envMigrationUser := "FLYWAY_USER"
	envMigrationPW := "FLYWAY_PASSWORD"
	jobName := "cockroachdb-cluster-migration"
	createFile := "create.sql"
	grantFile := "grant.sql"
	deleteFile := "delete.sql"

	destroyCM, err := configmap.AdaptFuncToDestroy(migrationConfigmap, namespace)
	if err != nil {
		return nil, nil, nil, err
	}

	destroyJ, err := job.AdaptFuncToDestroy(jobName, namespace)
	if err != nil {
		return nil, nil, nil, err
	}

	destroyers := []zitadel.DestroyFunc{
		zitadel.ResourceDestroyToZitadelDestroy(destroyJ),
		zitadel.ResourceDestroyToZitadelDestroy(destroyCM),
	}

	return func(k8sClient *kubernetes.Client, queried map[string]interface{}) (zitadel.EnsureFunc, error) {
			currentDB, err := coredb.ParseQueriedForDatabase(queried)
			if err != nil {
				return nil, err
			}

			dbHost := currentDB.GetURL()
			dbPort := currentDB.GetPort()

			internalLabels := make(map[string]string, 0)
			for k, v := range labels {
				internalLabels[k] = v
			}
			internalLabels["app.kubernetes.io/component"] = "migration"

			allScripts := scripts.GetAll()
			hash, err := getHash(allScripts)
			if err != nil {
				return nil, err
			}

			jobDef := &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      jobName,
					Namespace: namespace,
					Labels:    internalLabels,
					Annotations: map[string]string{
						"migrationhash": hash,
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
										"until pg_isready -h " + dbHost + " -p " + dbPort + "; do echo waiting for database; sleep 2; done;",
									},
								},
								{
									Name:  "create-flyway-user",
									Image: "cockroachdb/cockroach:v20.1.4",
									Env:   baseEnvVars(envMigrationUser, envMigrationPW, migrationUser, secretPasswordName),
									VolumeMounts: []corev1.VolumeMount{{
										Name:      rootUserInternal,
										MountPath: rootUserPath,
									}},
									Command: []string{"/bin/bash", "-c", "--"},
									Args: []string{
										strings.Join([]string{
											createUserCommand(envMigrationUser, envMigrationPW, createFile),
											grantUserCommand(envMigrationUser, grantFile),
											"cockroach.sh sql --certs-dir=/certificates --host=" + dbHost + ":" + dbPort + " -e \"$(cat " + createFile + ")\" -e \"$(cat " + grantFile + ")\";",
										},
											";"),
									},
								},
								{
									Name:            "db-migration",
									Image:           "flyway/flyway:6.5.0",
									ImagePullPolicy: "Always",
									Args: []string{
										"-url=jdbc:postgresql://" + dbHost + ":" + dbPort + "/defaultdb?&sslmode=verify-full&ssl=true&sslrootcert=" + rootUserPath + "/ca.crt&sslfactory=org.postgresql.ssl.NonValidatingFactory",
										"-locations=filesystem:" + migrationsPath,
										"migrate",
									},
									Env: migrationEnvVars(envMigrationUser, envMigrationPW, migrationUser, secretPasswordName, users),
									VolumeMounts: []corev1.VolumeMount{{
										Name:      migrationConfigmap,
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
									Image:   "cockroachdb/cockroach:v20.1.4",
									Command: []string{"/bin/bash", "-c", "--"},
									Args: []string{
										strings.Join([]string{
											deleteUserCommand(envMigrationUser, deleteFile),
											"cockroach.sh sql --certs-dir=/certificates --host=" + dbHost + ":" + dbPort + " -e \"$(cat " + deleteFile + ")\";",
										}, ";"),
									},
									Env: baseEnvVars(envMigrationUser, envMigrationPW, migrationUser, secretPasswordName),
									VolumeMounts: []corev1.VolumeMount{{
										Name:      rootUserInternal,
										MountPath: rootUserPath,
									}},
								},
							},
							RestartPolicy: "Never",
							Volumes: []corev1.Volume{{
								Name: migrationConfigmap,
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

			queryCM, err := configmap.AdaptFuncToEnsure(migrationConfigmap, namespace, labels, allScripts)
			if err != nil {
				return nil, err
			}
			queryJ, err := job.AdaptFuncToEnsure(jobDef)
			if err != nil {
				return nil, err
			}

			queriers := []zitadel.QueryFunc{
				zitadel.EnsureFuncToQueryFunc(currentDB.GetReadyQuery()),
				zitadel.ResourceQueryToZitadelQuery(queryCM),
				zitadel.ResourceQueryToZitadelQuery(queryJ),
			}
			return zitadel.QueriersToEnsureFunc(internalMonitor, true, queriers, k8sClient, queried)
		},
		zitadel.DestroyersToDestroyFunc(internalMonitor, destroyers),
		func(k8sClient *kubernetes.Client) error {
			internalMonitor.Info("waiting for migration to be completed")
			if err := k8sClient.WaitUntilJobCompleted(namespace, jobName, 60); err != nil {
				internalMonitor.Error(errors.Wrap(err, "error while waiting for migration to be completed"))
				return err
			}
			internalMonitor.Info("migration is completed, cleanup")
			if err := k8sClient.DeleteJob(namespace, jobName); err != nil {
				internalMonitor.Error(errors.Wrap(err, "error while trying to cleanup migration"))
				return err
			}
			internalMonitor.Info("migration cleanup is completed")
			return nil
		},
		nil
}

func createUserCommand(user, pw, file string) string {
	return strings.Join([]string{
		"echo -n 'CREATE USER IF NOT EXISTS ' > " + file,
		"echo -n ${" + user + "} >> " + file,
		"echo -n ';' >> " + file,
		"echo -n 'ALTER USER ' >> " + file,
		"echo -n ${" + user + "} >> " + file,
		"echo -n ' WITH PASSWORD ' >> " + file,
		"echo -n ${" + pw + "} >> " + file,
		"echo -n ';' >> " + file,
	}, ";")
}

func grantUserCommand(user, file string) string {
	return strings.Join([]string{
		"echo -n 'GRANT admin TO ' > " + file,
		"echo -n ${" + user + "} >> " + file,
		"echo -n ' WITH ADMIN OPTION;'  >> " + file,
	}, ";")
}
func deleteUserCommand(user, file string) string {
	return strings.Join([]string{
		"echo -n 'DROP USER IF EXISTS ' > " + file,
		"echo -n ${" + user + "} >> " + file,
		"echo -n ';' >> " + file,
	}, ";")
}

func baseEnvVars(envMigrationUser, envMigrationPW, migrationUser, userPasswordsSecret string) []corev1.EnvVar {
	envVars := []corev1.EnvVar{
		{
			Name:  envMigrationUser,
			Value: migrationUser,
		}, {
			Name: envMigrationPW,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: userPasswordsSecret},
					Key:                  migrationUser,
				},
			},
		},
	}
	return envVars
}

func migrationEnvVars(envMigrationUser, envMigrationPW, migrationUser, userPasswordsSecret string, users []string) []corev1.EnvVar {
	envVars := baseEnvVars(envMigrationUser, envMigrationPW, migrationUser, userPasswordsSecret)

	migrationEnvVars := make([]corev1.EnvVar, 0)
	for _, v := range envVars {
		migrationEnvVars = append(migrationEnvVars, v)
	}
	for _, user := range users {
		migrationEnvVars = append(migrationEnvVars, corev1.EnvVar{
			Name: "FLYWAY_PLACEHOLDERS_" + strings.ToUpper(user) + "PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: userPasswordsSecret},
					Key:                  migrationUser,
				},
			},
		})
	}
	return migrationEnvVars
}

func getHash(values map[string]string) (string, error) {
	scriptsStr := ""
	for k, v := range values {
		if scriptsStr == "" {
			scriptsStr = k + ": " + v
		} else {
			scriptsStr = scriptsStr + "," + k + ": " + v
		}
	}

	h := sha1.New()
	_, err := h.Write([]byte(scriptsStr))
	if err != nil {
		return "", err
	}
	hash := h.Sum(nil)
	return string(hash), err
}
