package managed

import (
	"github.com/caos/orbos/internal/operator/database/kinds/backups/bucket"
	"github.com/caos/orbos/internal/operator/database/kinds/backups/bucket/backup"
	"github.com/caos/orbos/internal/operator/database/kinds/backups/bucket/clean"
	"github.com/caos/orbos/internal/operator/database/kinds/backups/bucket/restore"
	"github.com/caos/orbos/mntr"
	kubernetesmock "github.com/caos/orbos/pkg/kubernetes/mock"
	"github.com/caos/orbos/pkg/secret"
	"github.com/caos/orbos/pkg/tree"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"testing"
	"time"
)

func getTreeWithDBAndBackup(t *testing.T, masterkey string, saJson string, backupName string) *tree.Tree {

	bucketDesired := getDesiredTree(t, masterkey, &bucket.DesiredV0{
		Common: &tree.Common{
			Kind:    "databases.caos.ch/BucketBackup",
			Version: "v0",
		},
		Spec: &bucket.Spec{
			Verbose: true,
			Cron:    "testCron",
			Bucket:  "testBucket",
			ServiceAccountJSON: &secret.Secret{
				Value: saJson,
			},
		},
	})
	bucketDesiredKind, err := bucket.ParseDesiredV0(bucketDesired)
	assert.NoError(t, err)
	bucketDesired.Parsed = bucketDesiredKind

	return getDesiredTree(t, masterkey, &DesiredV0{
		Common: &tree.Common{
			Kind:    "databases.caos.ch/CockroachDB",
			Version: "v0",
		},
		Spec: Spec{
			Verbose:         false,
			ReplicaCount:    1,
			StorageCapacity: "368Gi",
			StorageClass:    "testSC",
			NodeSelector:    map[string]string{},
			ClusterDns:      "testDns",
			Backups:         map[string]*tree.Tree{backupName: bucketDesired},
		},
	})
}

func TestManaged_AdaptBucketBackup(t *testing.T) {
	monitor := mntr.Monitor{}
	labels := map[string]string{"test": "test"}
	cockroachLabels := map[string]string{"test": "test", "app.kubernetes.io/component": "cockroachdb"}
	namespace := "testNs"
	timestamp := "testTs"
	nodeselector := map[string]string{"test": "test"}
	tolerations := []corev1.Toleration{}
	version := "testVersion"
	k8sClient := kubernetesmock.NewMockClientInt(gomock.NewController(t))
	backupName := "testBucket"
	saJson := "testSA"
	masterkey := "testMk"

	desired := getTreeWithDBAndBackup(t, masterkey, saJson, backupName)

	features := []string{backup.Normal}
	bucket.SetBackup(k8sClient, namespace, cockroachLabels, saJson)
	k8sClient.EXPECT().WaitUntilStatefulsetIsReady(namespace, sfsName, true, true, time.Duration(60))

	query, _, _, err := AdaptFunc(labels, namespace, timestamp, nodeselector, tolerations, version, features)(monitor, desired, &tree.Tree{})
	assert.NoError(t, err)

	databases := []string{"test1", "test2"}
	queried := bucket.SetQueriedForDatabases(databases)
	ensure, err := query(k8sClient, queried)
	assert.NoError(t, err)
	assert.NotNil(t, ensure)

	assert.NoError(t, ensure(k8sClient))
}

func TestManaged_AdaptBucketInstantBackup(t *testing.T) {
	monitor := mntr.Monitor{}
	labels := map[string]string{"test": "test"}
	cockroachLabels := map[string]string{"test": "test", "app.kubernetes.io/component": "cockroachdb"}
	namespace := "testNs"
	timestamp := "testTs"
	nodeselector := map[string]string{"test": "test"}
	tolerations := []corev1.Toleration{}
	version := "testVersion"
	masterkey := "testMk"
	k8sClient := kubernetesmock.NewMockClientInt(gomock.NewController(t))
	saJson := "testSA"
	backupName := "testBucket"

	features := []string{backup.Instant}
	bucket.SetInstantBackup(k8sClient, namespace, backupName, cockroachLabels, saJson)
	k8sClient.EXPECT().WaitUntilStatefulsetIsReady(namespace, sfsName, true, true, time.Duration(60))

	desired := getTreeWithDBAndBackup(t, masterkey, saJson, backupName)

	query, _, _, err := AdaptFunc(labels, namespace, timestamp, nodeselector, tolerations, version, features)(monitor, desired, &tree.Tree{})
	assert.NoError(t, err)

	databases := []string{"test1", "test2"}
	queried := bucket.SetQueriedForDatabases(databases)
	ensure, err := query(k8sClient, queried)
	assert.NoError(t, err)
	assert.NotNil(t, ensure)

	assert.NoError(t, ensure(k8sClient))
}

func TestManaged_AdaptBucketCleanAndRestore(t *testing.T) {
	monitor := mntr.Monitor{}
	labels := map[string]string{"test": "test"}
	cockroachLabels := map[string]string{"test": "test", "app.kubernetes.io/component": "cockroachdb"}
	namespace := "testNs"
	timestamp := "testTs"
	nodeselector := map[string]string{"test": "test"}
	tolerations := []corev1.Toleration{}
	version := "testVersion"
	masterkey := "testMk"
	k8sClient := kubernetesmock.NewMockClientInt(gomock.NewController(t))
	saJson := "testSA"
	backupName := "testBucket"

	features := []string{restore.Instant, clean.Instant}
	bucket.SetRestore(k8sClient, namespace, backupName, cockroachLabels, saJson)
	bucket.SetClean(k8sClient, namespace, backupName)
	k8sClient.EXPECT().WaitUntilStatefulsetIsReady(namespace, sfsName, true, true, time.Duration(60)).Times(2)

	desired := getTreeWithDBAndBackup(t, masterkey, saJson, backupName)

	query, _, _, err := AdaptFunc(labels, namespace, timestamp, nodeselector, tolerations, version, features)(monitor, desired, &tree.Tree{})
	assert.NoError(t, err)

	databases := []string{"test1", "test2"}
	queried := bucket.SetQueriedForDatabases(databases)
	ensure, err := query(k8sClient, queried)
	assert.NoError(t, err)
	assert.NotNil(t, ensure)

	assert.NoError(t, ensure(k8sClient))
}
