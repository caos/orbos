package job

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/kubernetes/resources"
	batch "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	macherrs "k8s.io/apimachinery/pkg/api/errors"
)

func AdaptFuncToEnsure(job *batch.Job) (resources.QueryFunc, error) {
	return func(k8sClient kubernetes.ClientInt) (resources.EnsureFunc, error) {

		jobDef, err := k8sClient.GetJob(job.GetNamespace(), job.GetName())
		if err != nil && !macherrs.IsNotFound(err) {
			return nil, err
		} else if macherrs.IsNotFound(err) {
			return func(k8sClient kubernetes.ClientInt) error {
				return k8sClient.ApplyJob(job)
			}, nil
		}

		jobDry := *job
		if job.Spec.Selector == nil {
			jobDry.Spec.Selector = jobDef.Spec.Selector
		}
		if job.Spec.Template.Labels == nil {
			jobDry.Spec.Template.Labels = jobDef.Spec.Template.Labels
		}
		if err := k8sClient.ApplyJobDryRun(&jobDry); err != nil && !strings.Contains(err.Error(), "field is immutable") {
			return nil, err
		}

		if jobDef.Spec.Template.Spec.SecurityContext != nil && jobDry.Spec.Template.Spec.SecurityContext == nil {
			jobDry.Spec.Template.Spec.SecurityContext = &corev1.PodSecurityContext{}
		}

		if reflect.DeepEqual(jobDry.Spec.Template.Spec, jobDef.Spec.Template.Spec) &&
			fmt.Sprint(jobDry.Labels) == fmt.Sprint(jobDef.Labels) &&
			fmt.Sprint(jobDry.Annotations) == fmt.Sprint(jobDef.Annotations) {
			return func(k8sClient kubernetes.ClientInt) error {
				return nil
			}, nil
		}

		return func(k8sClient kubernetes.ClientInt) error {
			if err := k8sClient.ApplyJob(job); err != nil && strings.Contains(err.Error(), "field is immutable") {
				if err := k8sClient.DeleteJob(job.GetNamespace(), job.GetName()); err != nil {
					return err
				}
				time.Sleep(1 * time.Second)
				return k8sClient.ApplyJob(job)
			}
			return err
		}, nil
	}, nil
}

func AdaptFuncToDestroy(namespace, name string) (resources.DestroyFunc, error) {
	return func(client kubernetes.ClientInt) error {
		return client.DeleteJob(namespace, name)
	}, nil
}
