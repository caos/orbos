// Code generated by MockGen. DO NOT EDIT.
// Source: client.go

// Package kubernetesmock is a generated GoMock package.
package kubernetesmock

import (
	mntr "github.com/caos/orbos/mntr"
	gomock "github.com/golang/mock/gomock"
	io "io"
	v1 "k8s.io/api/apps/v1"
	v10 "k8s.io/api/batch/v1"
	v1beta1 "k8s.io/api/batch/v1beta1"
	v11 "k8s.io/api/core/v1"
	v1beta10 "k8s.io/api/extensions/v1beta1"
	v1beta11 "k8s.io/api/policy/v1beta1"
	v12 "k8s.io/api/rbac/v1"
	v1beta12 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	unstructured "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	reflect "reflect"
	time "time"
)

// MockNodeWithKubeadm is a mock of NodeWithKubeadm interface
type MockNodeWithKubeadm struct {
	ctrl     *gomock.Controller
	recorder *MockNodeWithKubeadmMockRecorder
}

// MockNodeWithKubeadmMockRecorder is the mock recorder for MockNodeWithKubeadm
type MockNodeWithKubeadmMockRecorder struct {
	mock *MockNodeWithKubeadm
}

// NewMockNodeWithKubeadm creates a new mock instance
func NewMockNodeWithKubeadm(ctrl *gomock.Controller) *MockNodeWithKubeadm {
	mock := &MockNodeWithKubeadm{ctrl: ctrl}
	mock.recorder = &MockNodeWithKubeadmMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockNodeWithKubeadm) EXPECT() *MockNodeWithKubeadmMockRecorder {
	return m.recorder
}

// Execute mocks base method
func (m *MockNodeWithKubeadm) Execute(stdin io.Reader, cmd string) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Execute", stdin, cmd)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Execute indicates an expected call of Execute
func (mr *MockNodeWithKubeadmMockRecorder) Execute(stdin, cmd interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Execute", reflect.TypeOf((*MockNodeWithKubeadm)(nil).Execute), stdin, cmd)
}

// MockClientInt is a mock of ClientInt interface
type MockClientInt struct {
	ctrl     *gomock.Controller
	recorder *MockClientIntMockRecorder
}

// MockClientIntMockRecorder is the mock recorder for MockClientInt
type MockClientIntMockRecorder struct {
	mock *MockClientInt
}

// NewMockClientInt creates a new mock instance
func NewMockClientInt(ctrl *gomock.Controller) *MockClientInt {
	mock := &MockClientInt{ctrl: ctrl}
	mock.recorder = &MockClientIntMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockClientInt) EXPECT() *MockClientIntMockRecorder {
	return m.recorder
}

// ApplyService mocks base method
func (m *MockClientInt) ApplyService(rsc *v11.Service) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ApplyService", rsc)
	ret0, _ := ret[0].(error)
	return ret0
}

// ApplyService indicates an expected call of ApplyService
func (mr *MockClientIntMockRecorder) ApplyService(rsc interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ApplyService", reflect.TypeOf((*MockClientInt)(nil).ApplyService), rsc)
}

// DeleteService mocks base method
func (m *MockClientInt) DeleteService(namespace, name string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteService", namespace, name)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteService indicates an expected call of DeleteService
func (mr *MockClientIntMockRecorder) DeleteService(namespace, name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteService", reflect.TypeOf((*MockClientInt)(nil).DeleteService), namespace, name)
}

// GetJob mocks base method
func (m *MockClientInt) GetJob(namespace, name string) (*v10.Job, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetJob", namespace, name)
	ret0, _ := ret[0].(*v10.Job)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetJob indicates an expected call of GetJob
func (mr *MockClientIntMockRecorder) GetJob(namespace, name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetJob", reflect.TypeOf((*MockClientInt)(nil).GetJob), namespace, name)
}

// ApplyJob mocks base method
func (m *MockClientInt) ApplyJob(rsc *v10.Job) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ApplyJob", rsc)
	ret0, _ := ret[0].(error)
	return ret0
}

// ApplyJob indicates an expected call of ApplyJob
func (mr *MockClientIntMockRecorder) ApplyJob(rsc interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ApplyJob", reflect.TypeOf((*MockClientInt)(nil).ApplyJob), rsc)
}

// ApplyJobDryRun mocks base method
func (m *MockClientInt) ApplyJobDryRun(rsc *v10.Job) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ApplyJobDryRun", rsc)
	ret0, _ := ret[0].(error)
	return ret0
}

// ApplyJobDryRun indicates an expected call of ApplyJobDryRun
func (mr *MockClientIntMockRecorder) ApplyJobDryRun(rsc interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ApplyJobDryRun", reflect.TypeOf((*MockClientInt)(nil).ApplyJobDryRun), rsc)
}

// DeleteJob mocks base method
func (m *MockClientInt) DeleteJob(namespace, name string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteJob", namespace, name)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteJob indicates an expected call of DeleteJob
func (mr *MockClientIntMockRecorder) DeleteJob(namespace, name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteJob", reflect.TypeOf((*MockClientInt)(nil).DeleteJob), namespace, name)
}

// WaitUntilJobCompleted mocks base method
func (m *MockClientInt) WaitUntilJobCompleted(namespace, name string, timeout time.Duration) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WaitUntilJobCompleted", namespace, name, timeout)
	ret0, _ := ret[0].(error)
	return ret0
}

// WaitUntilJobCompleted indicates an expected call of WaitUntilJobCompleted
func (mr *MockClientIntMockRecorder) WaitUntilJobCompleted(namespace, name, timeout interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WaitUntilJobCompleted", reflect.TypeOf((*MockClientInt)(nil).WaitUntilJobCompleted), namespace, name, timeout)
}

// ApplyServiceAccount mocks base method
func (m *MockClientInt) ApplyServiceAccount(rsc *v11.ServiceAccount) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ApplyServiceAccount", rsc)
	ret0, _ := ret[0].(error)
	return ret0
}

// ApplyServiceAccount indicates an expected call of ApplyServiceAccount
func (mr *MockClientIntMockRecorder) ApplyServiceAccount(rsc interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ApplyServiceAccount", reflect.TypeOf((*MockClientInt)(nil).ApplyServiceAccount), rsc)
}

// DeleteServiceAccount mocks base method
func (m *MockClientInt) DeleteServiceAccount(namespace, name string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteServiceAccount", namespace, name)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteServiceAccount indicates an expected call of DeleteServiceAccount
func (mr *MockClientIntMockRecorder) DeleteServiceAccount(namespace, name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteServiceAccount", reflect.TypeOf((*MockClientInt)(nil).DeleteServiceAccount), namespace, name)
}

// ApplyStatefulSet mocks base method
func (m *MockClientInt) ApplyStatefulSet(rsc *v1.StatefulSet, force bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ApplyStatefulSet", rsc, force)
	ret0, _ := ret[0].(error)
	return ret0
}

// ApplyStatefulSet indicates an expected call of ApplyStatefulSet
func (mr *MockClientIntMockRecorder) ApplyStatefulSet(rsc, force interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ApplyStatefulSet", reflect.TypeOf((*MockClientInt)(nil).ApplyStatefulSet), rsc, force)
}

// DeleteStatefulset mocks base method
func (m *MockClientInt) DeleteStatefulset(namespace, name string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteStatefulset", namespace, name)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteStatefulset indicates an expected call of DeleteStatefulset
func (mr *MockClientIntMockRecorder) DeleteStatefulset(namespace, name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteStatefulset", reflect.TypeOf((*MockClientInt)(nil).DeleteStatefulset), namespace, name)
}

// ScaleStatefulset mocks base method
func (m *MockClientInt) ScaleStatefulset(namespace, name string, replicaCount int) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ScaleStatefulset", namespace, name, replicaCount)
	ret0, _ := ret[0].(error)
	return ret0
}

// ScaleStatefulset indicates an expected call of ScaleStatefulset
func (mr *MockClientIntMockRecorder) ScaleStatefulset(namespace, name, replicaCount interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ScaleStatefulset", reflect.TypeOf((*MockClientInt)(nil).ScaleStatefulset), namespace, name, replicaCount)
}

// WaitUntilStatefulsetIsReady mocks base method
func (m *MockClientInt) WaitUntilStatefulsetIsReady(namespace, name string, containerCheck, readyCheck bool, timeout time.Duration) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WaitUntilStatefulsetIsReady", namespace, name, containerCheck, readyCheck, timeout)
	ret0, _ := ret[0].(error)
	return ret0
}

// WaitUntilStatefulsetIsReady indicates an expected call of WaitUntilStatefulsetIsReady
func (mr *MockClientIntMockRecorder) WaitUntilStatefulsetIsReady(namespace, name, containerCheck, readyCheck, timeout interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WaitUntilStatefulsetIsReady", reflect.TypeOf((*MockClientInt)(nil).WaitUntilStatefulsetIsReady), namespace, name, containerCheck, readyCheck, timeout)
}

// ExecInPodWithOutput mocks base method
func (m *MockClientInt) ExecInPodWithOutput(namespace, name, container, command string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ExecInPodWithOutput", namespace, name, container, command)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ExecInPodWithOutput indicates an expected call of ExecInPodWithOutput
func (mr *MockClientIntMockRecorder) ExecInPodWithOutput(namespace, name, container, command interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ExecInPodWithOutput", reflect.TypeOf((*MockClientInt)(nil).ExecInPodWithOutput), namespace, name, container, command)
}

// ExecInPod mocks base method
func (m *MockClientInt) ExecInPod(namespace, name, container, command string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ExecInPod", namespace, name, container, command)
	ret0, _ := ret[0].(error)
	return ret0
}

// ExecInPod indicates an expected call of ExecInPod
func (mr *MockClientIntMockRecorder) ExecInPod(namespace, name, container, command interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ExecInPod", reflect.TypeOf((*MockClientInt)(nil).ExecInPod), namespace, name, container, command)
}

// GetDeployment mocks base method
func (m *MockClientInt) GetDeployment(namespace, name string) (*v1.Deployment, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDeployment", namespace, name)
	ret0, _ := ret[0].(*v1.Deployment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDeployment indicates an expected call of GetDeployment
func (mr *MockClientIntMockRecorder) GetDeployment(namespace, name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDeployment", reflect.TypeOf((*MockClientInt)(nil).GetDeployment), namespace, name)
}

// ApplyDeployment mocks base method
func (m *MockClientInt) ApplyDeployment(rsc *v1.Deployment, force bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ApplyDeployment", rsc, force)
	ret0, _ := ret[0].(error)
	return ret0
}

// ApplyDeployment indicates an expected call of ApplyDeployment
func (mr *MockClientIntMockRecorder) ApplyDeployment(rsc, force interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ApplyDeployment", reflect.TypeOf((*MockClientInt)(nil).ApplyDeployment), rsc, force)
}

// DeleteDeployment mocks base method
func (m *MockClientInt) DeleteDeployment(namespace, name string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteDeployment", namespace, name)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteDeployment indicates an expected call of DeleteDeployment
func (mr *MockClientIntMockRecorder) DeleteDeployment(namespace, name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteDeployment", reflect.TypeOf((*MockClientInt)(nil).DeleteDeployment), namespace, name)
}

// PatchDeployment mocks base method
func (m *MockClientInt) PatchDeployment(namespace, name, data string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PatchDeployment", namespace, name, data)
	ret0, _ := ret[0].(error)
	return ret0
}

// PatchDeployment indicates an expected call of PatchDeployment
func (mr *MockClientIntMockRecorder) PatchDeployment(namespace, name, data interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PatchDeployment", reflect.TypeOf((*MockClientInt)(nil).PatchDeployment), namespace, name, data)
}

// WaitUntilDeploymentReady mocks base method
func (m *MockClientInt) WaitUntilDeploymentReady(namespace, name string, containerCheck, readyCheck bool, timeout time.Duration) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WaitUntilDeploymentReady", namespace, name, containerCheck, readyCheck, timeout)
	ret0, _ := ret[0].(error)
	return ret0
}

// WaitUntilDeploymentReady indicates an expected call of WaitUntilDeploymentReady
func (mr *MockClientIntMockRecorder) WaitUntilDeploymentReady(namespace, name, containerCheck, readyCheck, timeout interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WaitUntilDeploymentReady", reflect.TypeOf((*MockClientInt)(nil).WaitUntilDeploymentReady), namespace, name, containerCheck, readyCheck, timeout)
}

// ScaleDeployment mocks base method
func (m *MockClientInt) ScaleDeployment(namespace, name string, replicaCount int) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ScaleDeployment", namespace, name, replicaCount)
	ret0, _ := ret[0].(error)
	return ret0
}

// ScaleDeployment indicates an expected call of ScaleDeployment
func (mr *MockClientIntMockRecorder) ScaleDeployment(namespace, name, replicaCount interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ScaleDeployment", reflect.TypeOf((*MockClientInt)(nil).ScaleDeployment), namespace, name, replicaCount)
}

// ExecInPodOfDeployment mocks base method
func (m *MockClientInt) ExecInPodOfDeployment(namespace, name, container, command string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ExecInPodOfDeployment", namespace, name, container, command)
	ret0, _ := ret[0].(error)
	return ret0
}

// ExecInPodOfDeployment indicates an expected call of ExecInPodOfDeployment
func (mr *MockClientIntMockRecorder) ExecInPodOfDeployment(namespace, name, container, command interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ExecInPodOfDeployment", reflect.TypeOf((*MockClientInt)(nil).ExecInPodOfDeployment), namespace, name, container, command)
}

// CheckCRD mocks base method
func (m *MockClientInt) CheckCRD(name string) (*v1beta12.CustomResourceDefinition, bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckCRD", name)
	ret0, _ := ret[0].(*v1beta12.CustomResourceDefinition)
	ret1, _ := ret[1].(bool)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// CheckCRD indicates an expected call of CheckCRD
func (mr *MockClientIntMockRecorder) CheckCRD(name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckCRD", reflect.TypeOf((*MockClientInt)(nil).CheckCRD), name)
}

// GetNamespacedCRDResource mocks base method
func (m *MockClientInt) GetNamespacedCRDResource(group, version, kind, namespace, name string) (*unstructured.Unstructured, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNamespacedCRDResource", group, version, kind, namespace, name)
	ret0, _ := ret[0].(*unstructured.Unstructured)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetNamespacedCRDResource indicates an expected call of GetNamespacedCRDResource
func (mr *MockClientIntMockRecorder) GetNamespacedCRDResource(group, version, kind, namespace, name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNamespacedCRDResource", reflect.TypeOf((*MockClientInt)(nil).GetNamespacedCRDResource), group, version, kind, namespace, name)
}

// ApplyNamespacedCRDResource mocks base method
func (m *MockClientInt) ApplyNamespacedCRDResource(group, version, kind, namespace, name string, crd *unstructured.Unstructured) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ApplyNamespacedCRDResource", group, version, kind, namespace, name, crd)
	ret0, _ := ret[0].(error)
	return ret0
}

// ApplyNamespacedCRDResource indicates an expected call of ApplyNamespacedCRDResource
func (mr *MockClientIntMockRecorder) ApplyNamespacedCRDResource(group, version, kind, namespace, name, crd interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ApplyNamespacedCRDResource", reflect.TypeOf((*MockClientInt)(nil).ApplyNamespacedCRDResource), group, version, kind, namespace, name, crd)
}

// DeleteNamespacedCRDResource mocks base method
func (m *MockClientInt) DeleteNamespacedCRDResource(group, version, kind, namespace, name string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteNamespacedCRDResource", group, version, kind, namespace, name)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteNamespacedCRDResource indicates an expected call of DeleteNamespacedCRDResource
func (mr *MockClientIntMockRecorder) DeleteNamespacedCRDResource(group, version, kind, namespace, name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteNamespacedCRDResource", reflect.TypeOf((*MockClientInt)(nil).DeleteNamespacedCRDResource), group, version, kind, namespace, name)
}

// ApplyCRDResource mocks base method
func (m *MockClientInt) ApplyCRDResource(crd *unstructured.Unstructured) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ApplyCRDResource", crd)
	ret0, _ := ret[0].(error)
	return ret0
}

// ApplyCRDResource indicates an expected call of ApplyCRDResource
func (mr *MockClientIntMockRecorder) ApplyCRDResource(crd interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ApplyCRDResource", reflect.TypeOf((*MockClientInt)(nil).ApplyCRDResource), crd)
}

// DeleteCRDResource mocks base method
func (m *MockClientInt) DeleteCRDResource(group, version, kind, name string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteCRDResource", group, version, kind, name)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteCRDResource indicates an expected call of DeleteCRDResource
func (mr *MockClientIntMockRecorder) DeleteCRDResource(group, version, kind, name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteCRDResource", reflect.TypeOf((*MockClientInt)(nil).DeleteCRDResource), group, version, kind, name)
}

// ApplyCronJob mocks base method
func (m *MockClientInt) ApplyCronJob(rsc *v1beta1.CronJob) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ApplyCronJob", rsc)
	ret0, _ := ret[0].(error)
	return ret0
}

// ApplyCronJob indicates an expected call of ApplyCronJob
func (mr *MockClientIntMockRecorder) ApplyCronJob(rsc interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ApplyCronJob", reflect.TypeOf((*MockClientInt)(nil).ApplyCronJob), rsc)
}

// DeleteCronJob mocks base method
func (m *MockClientInt) DeleteCronJob(namespace, name string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteCronJob", namespace, name)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteCronJob indicates an expected call of DeleteCronJob
func (mr *MockClientIntMockRecorder) DeleteCronJob(namespace, name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteCronJob", reflect.TypeOf((*MockClientInt)(nil).DeleteCronJob), namespace, name)
}

// ListCronJobs mocks base method
func (m *MockClientInt) ListCronJobs(namespace string, labels map[string]string) (*v1beta1.CronJobList, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListCronJobs", namespace, labels)
	ret0, _ := ret[0].(*v1beta1.CronJobList)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListCronJobs indicates an expected call of ListCronJobs
func (mr *MockClientIntMockRecorder) ListCronJobs(namespace, labels interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListCronJobs", reflect.TypeOf((*MockClientInt)(nil).ListCronJobs), namespace, labels)
}

// ListSecrets mocks base method
func (m *MockClientInt) ListSecrets(namespace string, labels map[string]string) (*v11.SecretList, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListSecrets", namespace, labels)
	ret0, _ := ret[0].(*v11.SecretList)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListSecrets indicates an expected call of ListSecrets
func (mr *MockClientIntMockRecorder) ListSecrets(namespace, labels interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListSecrets", reflect.TypeOf((*MockClientInt)(nil).ListSecrets), namespace, labels)
}

// GetSecret mocks base method
func (m *MockClientInt) GetSecret(namespace, name string) (*v11.Secret, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSecret", namespace, name)
	ret0, _ := ret[0].(*v11.Secret)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSecret indicates an expected call of GetSecret
func (mr *MockClientIntMockRecorder) GetSecret(namespace, name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSecret", reflect.TypeOf((*MockClientInt)(nil).GetSecret), namespace, name)
}

// ApplySecret mocks base method
func (m *MockClientInt) ApplySecret(rsc *v11.Secret) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ApplySecret", rsc)
	ret0, _ := ret[0].(error)
	return ret0
}

// ApplySecret indicates an expected call of ApplySecret
func (mr *MockClientIntMockRecorder) ApplySecret(rsc interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ApplySecret", reflect.TypeOf((*MockClientInt)(nil).ApplySecret), rsc)
}

// DeleteSecret mocks base method
func (m *MockClientInt) DeleteSecret(namespace, name string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteSecret", namespace, name)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteSecret indicates an expected call of DeleteSecret
func (mr *MockClientIntMockRecorder) DeleteSecret(namespace, name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteSecret", reflect.TypeOf((*MockClientInt)(nil).DeleteSecret), namespace, name)
}

// WaitForSecret mocks base method
func (m *MockClientInt) WaitForSecret(namespace, name string, timeout time.Duration) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WaitForSecret", namespace, name, timeout)
	ret0, _ := ret[0].(error)
	return ret0
}

// WaitForSecret indicates an expected call of WaitForSecret
func (mr *MockClientIntMockRecorder) WaitForSecret(namespace, name, timeout interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WaitForSecret", reflect.TypeOf((*MockClientInt)(nil).WaitForSecret), namespace, name, timeout)
}

// GetConfigMap mocks base method
func (m *MockClientInt) GetConfigMap(namespace, name string) (*v11.ConfigMap, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetConfigMap", namespace, name)
	ret0, _ := ret[0].(*v11.ConfigMap)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetConfigMap indicates an expected call of GetConfigMap
func (mr *MockClientIntMockRecorder) GetConfigMap(namespace, name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetConfigMap", reflect.TypeOf((*MockClientInt)(nil).GetConfigMap), namespace, name)
}

// ApplyConfigmap mocks base method
func (m *MockClientInt) ApplyConfigmap(rsc *v11.ConfigMap) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ApplyConfigmap", rsc)
	ret0, _ := ret[0].(error)
	return ret0
}

// ApplyConfigmap indicates an expected call of ApplyConfigmap
func (mr *MockClientIntMockRecorder) ApplyConfigmap(rsc interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ApplyConfigmap", reflect.TypeOf((*MockClientInt)(nil).ApplyConfigmap), rsc)
}

// DeleteConfigmap mocks base method
func (m *MockClientInt) DeleteConfigmap(namespace, name string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteConfigmap", namespace, name)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteConfigmap indicates an expected call of DeleteConfigmap
func (mr *MockClientIntMockRecorder) DeleteConfigmap(namespace, name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteConfigmap", reflect.TypeOf((*MockClientInt)(nil).DeleteConfigmap), namespace, name)
}

// WaitForConfigMap mocks base method
func (m *MockClientInt) WaitForConfigMap(namespace, name string, timeout time.Duration) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WaitForConfigMap", namespace, name, timeout)
	ret0, _ := ret[0].(error)
	return ret0
}

// WaitForConfigMap indicates an expected call of WaitForConfigMap
func (mr *MockClientIntMockRecorder) WaitForConfigMap(namespace, name, timeout interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WaitForConfigMap", reflect.TypeOf((*MockClientInt)(nil).WaitForConfigMap), namespace, name, timeout)
}

// ApplyRole mocks base method
func (m *MockClientInt) ApplyRole(rsc *v12.Role) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ApplyRole", rsc)
	ret0, _ := ret[0].(error)
	return ret0
}

// ApplyRole indicates an expected call of ApplyRole
func (mr *MockClientIntMockRecorder) ApplyRole(rsc interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ApplyRole", reflect.TypeOf((*MockClientInt)(nil).ApplyRole), rsc)
}

// DeleteRole mocks base method
func (m *MockClientInt) DeleteRole(namespace, name string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteRole", namespace, name)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteRole indicates an expected call of DeleteRole
func (mr *MockClientIntMockRecorder) DeleteRole(namespace, name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteRole", reflect.TypeOf((*MockClientInt)(nil).DeleteRole), namespace, name)
}

// ApplyClusterRole mocks base method
func (m *MockClientInt) ApplyClusterRole(rsc *v12.ClusterRole) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ApplyClusterRole", rsc)
	ret0, _ := ret[0].(error)
	return ret0
}

// ApplyClusterRole indicates an expected call of ApplyClusterRole
func (mr *MockClientIntMockRecorder) ApplyClusterRole(rsc interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ApplyClusterRole", reflect.TypeOf((*MockClientInt)(nil).ApplyClusterRole), rsc)
}

// DeleteClusterRole mocks base method
func (m *MockClientInt) DeleteClusterRole(name string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteClusterRole", name)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteClusterRole indicates an expected call of DeleteClusterRole
func (mr *MockClientIntMockRecorder) DeleteClusterRole(name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteClusterRole", reflect.TypeOf((*MockClientInt)(nil).DeleteClusterRole), name)
}

// ApplyIngress mocks base method
func (m *MockClientInt) ApplyIngress(rsc *v1beta10.Ingress) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ApplyIngress", rsc)
	ret0, _ := ret[0].(error)
	return ret0
}

// ApplyIngress indicates an expected call of ApplyIngress
func (mr *MockClientIntMockRecorder) ApplyIngress(rsc interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ApplyIngress", reflect.TypeOf((*MockClientInt)(nil).ApplyIngress), rsc)
}

// DeleteIngress mocks base method
func (m *MockClientInt) DeleteIngress(namespace, name string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteIngress", namespace, name)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteIngress indicates an expected call of DeleteIngress
func (mr *MockClientIntMockRecorder) DeleteIngress(namespace, name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteIngress", reflect.TypeOf((*MockClientInt)(nil).DeleteIngress), namespace, name)
}

// ApplyRoleBinding mocks base method
func (m *MockClientInt) ApplyRoleBinding(rsc *v12.RoleBinding) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ApplyRoleBinding", rsc)
	ret0, _ := ret[0].(error)
	return ret0
}

// ApplyRoleBinding indicates an expected call of ApplyRoleBinding
func (mr *MockClientIntMockRecorder) ApplyRoleBinding(rsc interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ApplyRoleBinding", reflect.TypeOf((*MockClientInt)(nil).ApplyRoleBinding), rsc)
}

// DeleteRoleBinding mocks base method
func (m *MockClientInt) DeleteRoleBinding(namespace, name string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteRoleBinding", namespace, name)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteRoleBinding indicates an expected call of DeleteRoleBinding
func (mr *MockClientIntMockRecorder) DeleteRoleBinding(namespace, name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteRoleBinding", reflect.TypeOf((*MockClientInt)(nil).DeleteRoleBinding), namespace, name)
}

// ApplyClusterRoleBinding mocks base method
func (m *MockClientInt) ApplyClusterRoleBinding(rsc *v12.ClusterRoleBinding) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ApplyClusterRoleBinding", rsc)
	ret0, _ := ret[0].(error)
	return ret0
}

// ApplyClusterRoleBinding indicates an expected call of ApplyClusterRoleBinding
func (mr *MockClientIntMockRecorder) ApplyClusterRoleBinding(rsc interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ApplyClusterRoleBinding", reflect.TypeOf((*MockClientInt)(nil).ApplyClusterRoleBinding), rsc)
}

// DeleteClusterRoleBinding mocks base method
func (m *MockClientInt) DeleteClusterRoleBinding(name string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteClusterRoleBinding", name)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteClusterRoleBinding indicates an expected call of DeleteClusterRoleBinding
func (mr *MockClientIntMockRecorder) DeleteClusterRoleBinding(name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteClusterRoleBinding", reflect.TypeOf((*MockClientInt)(nil).DeleteClusterRoleBinding), name)
}

// ApplyPodDisruptionBudget mocks base method
func (m *MockClientInt) ApplyPodDisruptionBudget(rsc *v1beta11.PodDisruptionBudget) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ApplyPodDisruptionBudget", rsc)
	ret0, _ := ret[0].(error)
	return ret0
}

// ApplyPodDisruptionBudget indicates an expected call of ApplyPodDisruptionBudget
func (mr *MockClientIntMockRecorder) ApplyPodDisruptionBudget(rsc interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ApplyPodDisruptionBudget", reflect.TypeOf((*MockClientInt)(nil).ApplyPodDisruptionBudget), rsc)
}

// DeletePodDisruptionBudget mocks base method
func (m *MockClientInt) DeletePodDisruptionBudget(namespace, name string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeletePodDisruptionBudget", namespace, name)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeletePodDisruptionBudget indicates an expected call of DeletePodDisruptionBudget
func (mr *MockClientIntMockRecorder) DeletePodDisruptionBudget(namespace, name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeletePodDisruptionBudget", reflect.TypeOf((*MockClientInt)(nil).DeletePodDisruptionBudget), namespace, name)
}

// ApplyNamespace mocks base method
func (m *MockClientInt) ApplyNamespace(rsc *v11.Namespace) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ApplyNamespace", rsc)
	ret0, _ := ret[0].(error)
	return ret0
}

// ApplyNamespace indicates an expected call of ApplyNamespace
func (mr *MockClientIntMockRecorder) ApplyNamespace(rsc interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ApplyNamespace", reflect.TypeOf((*MockClientInt)(nil).ApplyNamespace), rsc)
}

// DeleteNamespace mocks base method
func (m *MockClientInt) DeleteNamespace(name string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteNamespace", name)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteNamespace indicates an expected call of DeleteNamespace
func (mr *MockClientIntMockRecorder) DeleteNamespace(name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteNamespace", reflect.TypeOf((*MockClientInt)(nil).DeleteNamespace), name)
}

// ListPersistentVolumes mocks base method
func (m *MockClientInt) ListPersistentVolumes() (*v11.PersistentVolumeList, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListPersistentVolumes")
	ret0, _ := ret[0].(*v11.PersistentVolumeList)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListPersistentVolumes indicates an expected call of ListPersistentVolumes
func (mr *MockClientIntMockRecorder) ListPersistentVolumes() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListPersistentVolumes", reflect.TypeOf((*MockClientInt)(nil).ListPersistentVolumes))
}

// ApplyPlainYAML mocks base method
func (m *MockClientInt) ApplyPlainYAML(arg0 mntr.Monitor, arg1 []byte) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ApplyPlainYAML", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// ApplyPlainYAML indicates an expected call of ApplyPlainYAML
func (mr *MockClientIntMockRecorder) ApplyPlainYAML(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ApplyPlainYAML", reflect.TypeOf((*MockClientInt)(nil).ApplyPlainYAML), arg0, arg1)
}

// ListPersistentVolumeClaims mocks base method
func (m *MockClientInt) ListPersistentVolumeClaims(namespace string) (*v11.PersistentVolumeClaimList, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListPersistentVolumeClaims", namespace)
	ret0, _ := ret[0].(*v11.PersistentVolumeClaimList)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListPersistentVolumeClaims indicates an expected call of ListPersistentVolumeClaims
func (mr *MockClientIntMockRecorder) ListPersistentVolumeClaims(namespace interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListPersistentVolumeClaims", reflect.TypeOf((*MockClientInt)(nil).ListPersistentVolumeClaims), namespace)
}

// DeletePersistentVolumeClaim mocks base method
func (m *MockClientInt) DeletePersistentVolumeClaim(namespace, name string, timeout time.Duration) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeletePersistentVolumeClaim", namespace, name, timeout)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeletePersistentVolumeClaim indicates an expected call of DeletePersistentVolumeClaim
func (mr *MockClientIntMockRecorder) DeletePersistentVolumeClaim(namespace, name, timeout interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeletePersistentVolumeClaim", reflect.TypeOf((*MockClientInt)(nil).DeletePersistentVolumeClaim), namespace, name, timeout)
}

// MockMachine is a mock of Machine interface
type MockMachine struct {
	ctrl     *gomock.Controller
	recorder *MockMachineMockRecorder
}

// MockMachineMockRecorder is the mock recorder for MockMachine
type MockMachineMockRecorder struct {
	mock *MockMachine
}

// NewMockMachine creates a new mock instance
func NewMockMachine(ctrl *gomock.Controller) *MockMachine {
	mock := &MockMachine{ctrl: ctrl}
	mock.recorder = &MockMachineMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockMachine) EXPECT() *MockMachineMockRecorder {
	return m.recorder
}

// GetUpdating mocks base method
func (m *MockMachine) GetUpdating() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUpdating")
	ret0, _ := ret[0].(bool)
	return ret0
}

// GetUpdating indicates an expected call of GetUpdating
func (mr *MockMachineMockRecorder) GetUpdating() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUpdating", reflect.TypeOf((*MockMachine)(nil).GetUpdating))
}

// SetUpdating mocks base method
func (m *MockMachine) SetUpdating(arg0 bool) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetUpdating", arg0)
}

// SetUpdating indicates an expected call of SetUpdating
func (mr *MockMachineMockRecorder) SetUpdating(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetUpdating", reflect.TypeOf((*MockMachine)(nil).SetUpdating), arg0)
}

// GetJoined mocks base method
func (m *MockMachine) GetJoined() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetJoined")
	ret0, _ := ret[0].(bool)
	return ret0
}

// GetJoined indicates an expected call of GetJoined
func (mr *MockMachineMockRecorder) GetJoined() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetJoined", reflect.TypeOf((*MockMachine)(nil).GetJoined))
}

// SetJoined mocks base method
func (m *MockMachine) SetJoined(arg0 bool) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetJoined", arg0)
}

// SetJoined indicates an expected call of SetJoined
func (mr *MockMachineMockRecorder) SetJoined(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetJoined", reflect.TypeOf((*MockMachine)(nil).SetJoined), arg0)
}
