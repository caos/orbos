package bucket

import (
	"github.com/caos/orbos/internal/operator/database/kinds/backups/bucket/backup"
	"github.com/caos/orbos/internal/operator/database/kinds/backups/bucket/clean"
	"github.com/caos/orbos/internal/operator/database/kinds/backups/bucket/restore"
	"github.com/caos/orbos/internal/operator/database/kinds/databases/core"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/kubernetes"
	kubernetesmock "github.com/caos/orbos/pkg/kubernetes/mock"
	"github.com/caos/orbos/pkg/secret"
	"github.com/caos/orbos/pkg/tree"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	macherrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"testing"
)

func SetQueriedForDatabases(databases []string) map[string]interface{} {
	queried := map[string]interface{}{}
	core.SetQueriedForDatabaseDBList(queried, databases)

	return queried
}

func SetInstantBackup(
	k8sClient *kubernetesmock.MockClientInt,
	namespace string,
	backupName string,
	labels map[string]string,
	saJson string,
) {

	k8sClient.EXPECT().ApplySecret(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
			Labels:    labels,
		},
		StringData: map[string]string{secretKey: saJson},
		Type:       "Opaque",
	}).Times(1).Return(nil)

	k8sClient.EXPECT().ApplyJob(gomock.Any()).Times(1).Return(nil)
	k8sClient.EXPECT().GetJob(namespace, backup.GetJobName(backupName)).Times(1).Return(nil, macherrs.NewNotFound(schema.GroupResource{"batch", "jobs"}, backup.GetJobName(backupName)))
	k8sClient.EXPECT().WaitUntilJobCompleted(namespace, backup.GetJobName(backupName), gomock.Any()).Times(1).Return(nil)
	k8sClient.EXPECT().DeleteJob(namespace, backup.GetJobName(backupName)).Times(1).Return(nil)
}

func SetBackup(
	k8sClient *kubernetesmock.MockClientInt,
	namespace string,
	labels map[string]string,
	saJson string,
) {
	k8sClient.EXPECT().ApplySecret(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
			Labels:    labels,
		},
		StringData: map[string]string{secretKey: saJson},
		Type:       "Opaque",
	}).Times(1).Return(nil)
	k8sClient.EXPECT().ApplyCronJob(gomock.Any()).Times(1).Return(nil)
}

func SetClean(
	k8sClient *kubernetesmock.MockClientInt,
	namespace string,
	backupName string,
) {

	k8sClient.EXPECT().ApplyJob(gomock.Any()).Times(1).Return(nil)
	k8sClient.EXPECT().GetJob(namespace, clean.GetJobName(backupName)).Times(1).Return(nil, macherrs.NewNotFound(schema.GroupResource{"batch", "jobs"}, clean.GetJobName(backupName)))
	k8sClient.EXPECT().WaitUntilJobCompleted(namespace, clean.GetJobName(backupName), gomock.Any()).Times(1).Return(nil)
	k8sClient.EXPECT().DeleteJob(namespace, clean.GetJobName(backupName)).Times(1).Return(nil)
}

func SetRestore(
	k8sClient *kubernetesmock.MockClientInt,
	namespace string,
	backupName string,
	labels map[string]string,
	saJson string,
) {
	k8sClient.EXPECT().ApplySecret(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
			Labels:    labels,
		},
		StringData: map[string]string{secretKey: saJson},
		Type:       "Opaque",
	}).Times(1).Return(nil)

	k8sClient.EXPECT().ApplyJob(gomock.Any()).Times(1).Return(nil)
	k8sClient.EXPECT().GetJob(namespace, restore.GetJobName(backupName)).Times(1).Return(nil, macherrs.NewNotFound(schema.GroupResource{"batch", "jobs"}, restore.GetJobName(backupName)))
	k8sClient.EXPECT().WaitUntilJobCompleted(namespace, restore.GetJobName(backupName), gomock.Any()).Times(1).Return(nil)
	k8sClient.EXPECT().DeleteJob(namespace, restore.GetJobName(backupName)).Times(1).Return(nil)
}

func getDesiredTree(t *testing.T, masterkey string, desired *DesiredV0) *tree.Tree {
	secret.Masterkey = masterkey

	desiredTree := &tree.Tree{}
	data, err := yaml.Marshal(desired)
	assert.NoError(t, err)
	assert.NoError(t, yaml.Unmarshal(data, desiredTree))

	return desiredTree
}

func TestBucket_Secrets(t *testing.T) {
	masterkey := "testMk"
	features := []string{backup.Normal}
	saJson := "testSA"

	bucketName := "testBucket2"
	cron := "testCron2"
	monitor := mntr.Monitor{}
	namespace := "testNs2"
	labels := map[string]string{"test2": "test2"}
	timestamp := "test2"
	nodeselector := map[string]string{"test2": "test2"}
	tolerations := []corev1.Toleration{
		{Key: "testKey2", Operator: "testOp2"}}
	backupName := "testName2"
	version := "testVersion2"

	desired := getDesiredTree(t, masterkey, &DesiredV0{
		Common: &tree.Common{
			Kind:    "databases.caos.ch/BucketBackup",
			Version: "v0",
		},
		Spec: &Spec{
			Verbose: true,
			Cron:    cron,
			Bucket:  bucketName,
			ServiceAccountJSON: &secret.Secret{
				Value: saJson,
			},
		},
	})

	checkDBReady := func(k8sClient kubernetes.ClientInt) error {
		return nil
	}

	allSecrets := map[string]string{
		"serviceaccountjson": saJson,
	}

	_, _, secrets, err := AdaptFunc(
		backupName,
		namespace,
		labels,
		checkDBReady,
		timestamp,
		nodeselector,
		tolerations,
		version,
		features,
	)(
		monitor,
		desired,
		&tree.Tree{},
	)
	assert.NoError(t, err)
	for key, value := range allSecrets {
		assert.Contains(t, secrets, key)
		assert.Equal(t, value, secrets[key].Value)
	}
}

func TestBucket_AdaptBackup(t *testing.T) {
	masterkey := "testMk"
	client := kubernetesmock.NewMockClientInt(gomock.NewController(t))
	features := []string{backup.Normal}
	saJson := "testSA"

	bucketName := "testBucket2"
	cron := "testCron2"
	monitor := mntr.Monitor{}
	namespace := "testNs2"
	labels := map[string]string{"test2": "test2"}
	timestamp := "test2"
	nodeselector := map[string]string{"test2": "test2"}
	tolerations := []corev1.Toleration{
		{Key: "testKey2", Operator: "testOp2"}}
	backupName := "testName2"
	version := "testVersion2"

	desired := getDesiredTree(t, masterkey, &DesiredV0{
		Common: &tree.Common{
			Kind:    "databases.caos.ch/BucketBackup",
			Version: "v0",
		},
		Spec: &Spec{
			Verbose: true,
			Cron:    cron,
			Bucket:  bucketName,
			ServiceAccountJSON: &secret.Secret{
				Value: saJson,
			},
		},
	})

	checkDBReady := func(k8sClient kubernetes.ClientInt) error {
		return nil
	}

	SetBackup(client, namespace, labels, saJson)

	query, _, _, err := AdaptFunc(
		backupName,
		namespace,
		labels,
		checkDBReady,
		timestamp,
		nodeselector,
		tolerations,
		version,
		features,
	)(
		monitor,
		desired,
		&tree.Tree{},
	)

	assert.NoError(t, err)
	databases := []string{"test1", "test2"}
	queried := SetQueriedForDatabases(databases)
	ensure, err := query(client, queried)
	assert.NoError(t, err)
	assert.NotNil(t, ensure)
	assert.NoError(t, ensure(client))
}

func TestBucket_AdaptInstantBackup(t *testing.T) {
	masterkey := "testMk"
	client := kubernetesmock.NewMockClientInt(gomock.NewController(t))
	features := []string{backup.Instant}

	bucketName := "testBucket1"
	cron := "testCron"
	monitor := mntr.Monitor{}
	namespace := "testNs"
	labels := map[string]string{"test": "test"}
	timestamp := "test"
	nodeselector := map[string]string{"test": "test"}
	tolerations := []corev1.Toleration{
		{Key: "testKey", Operator: "testOp"}}
	backupName := "testName"
	version := "testVersion"
	saJson := "testSA"

	desired := getDesiredTree(t, masterkey, &DesiredV0{
		Common: &tree.Common{
			Kind:    "databases.caos.ch/BucketBackup",
			Version: "v0",
		},
		Spec: &Spec{
			Verbose: true,
			Cron:    cron,
			Bucket:  bucketName,
			ServiceAccountJSON: &secret.Secret{
				Value: saJson,
			},
		},
	})

	checkDBReady := func(k8sClient kubernetes.ClientInt) error {
		return nil
	}

	SetInstantBackup(client, namespace, backupName, labels, saJson)

	query, _, _, err := AdaptFunc(
		backupName,
		namespace,
		labels,
		checkDBReady,
		timestamp,
		nodeselector,
		tolerations,
		version,
		features,
	)(
		monitor,
		desired,
		&tree.Tree{},
	)

	assert.NoError(t, err)
	databases := []string{"test1", "test2"}
	queried := SetQueriedForDatabases(databases)
	ensure, err := query(client, queried)
	assert.NotNil(t, ensure)
	assert.NoError(t, err)
	assert.NoError(t, ensure(client))
}

func TestBucket_AdaptRestore(t *testing.T) {
	masterkey := "testMk"
	client := kubernetesmock.NewMockClientInt(gomock.NewController(t))
	features := []string{restore.Instant}

	bucketName := "testBucket1"
	cron := "testCron"
	monitor := mntr.Monitor{}
	namespace := "testNs"
	labels := map[string]string{"test": "test"}
	timestamp := "test"
	nodeselector := map[string]string{"test": "test"}
	tolerations := []corev1.Toleration{
		{Key: "testKey", Operator: "testOp"}}
	backupName := "testName"
	version := "testVersion"
	saJson := "testSA"

	desired := getDesiredTree(t, masterkey, &DesiredV0{
		Common: &tree.Common{
			Kind:    "databases.caos.ch/BucketBackup",
			Version: "v0",
		},
		Spec: &Spec{
			Verbose: true,
			Cron:    cron,
			Bucket:  bucketName,
			ServiceAccountJSON: &secret.Secret{
				Value: saJson,
			},
		},
	})

	checkDBReady := func(k8sClient kubernetes.ClientInt) error {
		return nil
	}

	SetRestore(client, namespace, backupName, labels, saJson)

	query, _, _, err := AdaptFunc(
		backupName,
		namespace,
		labels,
		checkDBReady,
		timestamp,
		nodeselector,
		tolerations,
		version,
		features,
	)(
		monitor,
		desired,
		&tree.Tree{},
	)

	assert.NoError(t, err)
	databases := []string{"test1", "test2"}
	queried := SetQueriedForDatabases(databases)
	ensure, err := query(client, queried)
	assert.NotNil(t, ensure)
	assert.NoError(t, err)
	assert.NoError(t, ensure(client))
}

func TestBucket_AdaptClean(t *testing.T) {
	masterkey := "testMk"
	client := kubernetesmock.NewMockClientInt(gomock.NewController(t))
	features := []string{clean.Instant}

	bucketName := "testBucket1"
	cron := "testCron"
	monitor := mntr.Monitor{}
	namespace := "testNs"
	labels := map[string]string{"test": "test"}
	timestamp := "test"
	nodeselector := map[string]string{"test": "test"}
	tolerations := []corev1.Toleration{
		{Key: "testKey", Operator: "testOp"}}
	backupName := "testName"
	version := "testVersion"
	saJson := "testSA"

	desired := getDesiredTree(t, masterkey, &DesiredV0{
		Common: &tree.Common{
			Kind:    "databases.caos.ch/BucketBackup",
			Version: "v0",
		},
		Spec: &Spec{
			Verbose: true,
			Cron:    cron,
			Bucket:  bucketName,
			ServiceAccountJSON: &secret.Secret{
				Value: saJson,
			},
		},
	})

	checkDBReady := func(k8sClient kubernetes.ClientInt) error {
		return nil
	}

	SetClean(client, namespace, backupName)

	query, _, _, err := AdaptFunc(
		backupName,
		namespace,
		labels,
		checkDBReady,
		timestamp,
		nodeselector,
		tolerations,
		version,
		features,
	)(
		monitor,
		desired,
		&tree.Tree{},
	)

	assert.NoError(t, err)
	databases := []string{"test1", "test2"}
	queried := SetQueriedForDatabases(databases)
	ensure, err := query(client, queried)
	assert.NotNil(t, ensure)
	assert.NoError(t, err)
	assert.NoError(t, ensure(client))
}
