package restore

import (
	"github.com/caos/orbos/internal/utils/helper"
	"github.com/stretchr/testify/assert"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestBackup_Job1(t *testing.T) {
	nodeselector := map[string]string{"test": "test"}
	tolerations := []corev1.Toleration{
		{Key: "testKey", Operator: "testOp"}}
	version := "testVersion"
	command := "test"
	secretKey := "testKey"
	secretName := "testSecretName"
	jobName := "testJob"
	namespace := "testNs"
	labels := map[string]string{"test": "test"}

	equals :=
		&batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      jobName,
				Namespace: namespace,
				Labels:    labels,
			},
			Spec: batchv1.JobSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						RestartPolicy: corev1.RestartPolicyNever,
						NodeSelector:  nodeselector,
						Tolerations:   tolerations,
						Containers: []corev1.Container{{
							Name:  jobName,
							Image: image + ":" + version,
							Command: []string{
								"/bin/bash",
								"-c",
								command,
							},
							VolumeMounts: []corev1.VolumeMount{{
								Name:      internalSecretName,
								MountPath: certPath,
							}, {
								Name:      secretKey,
								SubPath:   secretKey,
								MountPath: secretPath,
							}},
							ImagePullPolicy: corev1.PullAlways,
						}},
						Volumes: []corev1.Volume{{
							Name: internalSecretName,
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName:  rootSecretName,
									DefaultMode: helper.PointerInt32(defaultMode),
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
					},
				},
			},
		}

	assert.Equal(t, equals, getJob(namespace, labels, jobName, nodeselector, tolerations, secretName, secretKey, version, command))
}

func TestBackup_Job2(t *testing.T) {
	nodeselector := map[string]string{"test2": "test2"}
	tolerations := []corev1.Toleration{
		{Key: "testKey2", Operator: "testOp2"}}
	version := "testVersion2"
	command := "test2"
	secretKey := "testKey2"
	secretName := "testSecretName2"
	jobName := "testJob2"
	namespace := "testNs2"
	labels := map[string]string{"test2": "test2"}

	equals :=
		&batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      jobName,
				Namespace: namespace,
				Labels:    labels,
			},
			Spec: batchv1.JobSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						RestartPolicy: corev1.RestartPolicyNever,
						NodeSelector:  nodeselector,
						Tolerations:   tolerations,
						Containers: []corev1.Container{{
							Name:  jobName,
							Image: image + ":" + version,
							Command: []string{
								"/bin/bash",
								"-c",
								command,
							},
							VolumeMounts: []corev1.VolumeMount{{
								Name:      internalSecretName,
								MountPath: certPath,
							}, {
								Name:      secretKey,
								SubPath:   secretKey,
								MountPath: secretPath,
							}},
							ImagePullPolicy: corev1.PullAlways,
						}},
						Volumes: []corev1.Volume{{
							Name: internalSecretName,
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName:  rootSecretName,
									DefaultMode: helper.PointerInt32(defaultMode),
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
					},
				},
			},
		}

	assert.Equal(t, equals, getJob(namespace, labels, jobName, nodeselector, tolerations, secretName, secretKey, version, command))
}
