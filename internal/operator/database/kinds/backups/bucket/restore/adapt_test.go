package restore

import (
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/kubernetes"
	kubernetesmock "github.com/caos/orbos/pkg/kubernetes/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	macherrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"testing"
)

func TestBackup_Adapt1(t *testing.T) {
	client := kubernetesmock.NewMockClientInt(gomock.NewController(t))

	monitor := mntr.Monitor{}
	namespace := "testNs"
	labels := map[string]string{"test": "test"}
	databases := []string{"testDb"}
	nodeselector := map[string]string{"test": "test"}
	tolerations := []corev1.Toleration{
		{Key: "testKey", Operator: "testOp"}}
	timestamp := "testTs"
	backupName := "testName2"
	bucketName := "testBucket2"
	version := "testVersion"
	secretKey := "testKey"
	secretName := "testSecretName"
	jobName := GetJobName(backupName)

	checkDBReady := func(k8sClient kubernetes.ClientInt) error {
		return nil
	}

	jobDef := getJob(
		namespace,
		labels,
		jobName,
		nodeselector,
		tolerations,
		secretName,
		secretKey,
		version,
		getCommand(
			timestamp,
			databases,
			bucketName,
			backupName,
		),
	)

	client.EXPECT().ApplyJob(jobDef).Times(1).Return(nil)
	client.EXPECT().GetJob(jobDef.Namespace, jobDef.Name).Times(1).Return(nil, macherrs.NewNotFound(schema.GroupResource{"batch", "jobs"}, jobName))
	client.EXPECT().WaitUntilJobCompleted(jobDef.Namespace, jobDef.Name, timeout).Times(1).Return(nil)
	client.EXPECT().DeleteJob(jobDef.Namespace, jobDef.Name).Times(1).Return(nil)

	query, _, err := AdaptFunc(
		monitor,
		backupName,
		namespace,
		labels,
		databases,
		bucketName,
		timestamp,
		nodeselector,
		tolerations,
		checkDBReady,
		secretName,
		secretKey,
		version,
	)

	assert.NoError(t, err)
	queried := map[string]interface{}{}
	ensure, err := query(client, queried)
	assert.NoError(t, err)
	assert.NoError(t, ensure(client))
}

func TestBackup_Adapt2(t *testing.T) {
	client := kubernetesmock.NewMockClientInt(gomock.NewController(t))

	monitor := mntr.Monitor{}
	namespace := "testNs2"
	labels := map[string]string{"test2": "test2"}
	databases := []string{"testDb1", "testDb2"}
	nodeselector := map[string]string{"test2": "test2"}
	tolerations := []corev1.Toleration{
		{Key: "testKey2", Operator: "testOp2"}}
	timestamp := "testTs"
	backupName := "testName2"
	bucketName := "testBucket2"
	version := "testVersion2"
	secretKey := "testKey2"
	secretName := "testSecretName2"
	jobName := GetJobName(backupName)

	checkDBReady := func(k8sClient kubernetes.ClientInt) error {
		return nil
	}

	jobDef := getJob(
		namespace,
		labels,
		jobName,
		nodeselector,
		tolerations,
		secretName,
		secretKey,
		version,
		getCommand(
			timestamp,
			databases,
			bucketName,
			backupName,
		),
	)

	client.EXPECT().ApplyJob(jobDef).Times(1).Return(nil)
	client.EXPECT().GetJob(jobDef.Namespace, jobDef.Name).Times(1).Return(nil, macherrs.NewNotFound(schema.GroupResource{"batch", "jobs"}, jobName))
	client.EXPECT().WaitUntilJobCompleted(jobDef.Namespace, jobDef.Name, timeout).Times(1).Return(nil)
	client.EXPECT().DeleteJob(jobDef.Namespace, jobDef.Name).Times(1).Return(nil)

	query, _, err := AdaptFunc(
		monitor,
		backupName,
		namespace,
		labels,
		databases,
		bucketName,
		timestamp,
		nodeselector,
		tolerations,
		checkDBReady,
		secretName,
		secretKey,
		version,
	)

	assert.NoError(t, err)
	queried := map[string]interface{}{}
	ensure, err := query(client, queried)
	assert.NoError(t, err)
	assert.NoError(t, ensure(client))
}
