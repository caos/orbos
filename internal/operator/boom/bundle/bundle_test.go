package bundle

import (
	"sync"

	"github.com/caos/orbos/internal/operator/boom/api/v1beta1"
	application "github.com/caos/orbos/internal/operator/boom/application/mock"
	"github.com/caos/orbos/internal/operator/boom/bundle/bundles"
	"github.com/caos/orbos/internal/operator/boom/bundle/config"
	"github.com/caos/orbos/internal/operator/boom/name"
	"github.com/caos/orbos/internal/operator/boom/templator/yaml"
	"github.com/caos/orbos/internal/utils/clientgo"
	"github.com/caos/orbos/internal/utils/helper"
	"github.com/caos/orbos/mntr"
	"github.com/stretchr/testify/assert"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	yamlv3 "gopkg.in/yaml.v3"

	"testing"
)

const (
	baseDirectoryPath       = "../../tools"
	dashboardsDirectoryPath = "../../dashboards"
)

var (
	testHelperResource = &helper.Resource{
		Kind:       "test",
		ApiVersion: "test/v1",
		Metadata: &helper.Metadata{
			Name:      "test",
			Namespace: "test",
		},
	}
	testClientgoResource = &clientgo.Resource{
		Group:     "test",
		Version:   "v1",
		Resource:  "test",
		Kind:      "test",
		Name:      "test",
		Namespace: "test",
		Labels:    map[string]string{"test": "test"},
	}
)

func newMonitor() mntr.Monitor {
	monitor := mntr.Monitor{
		OnInfo:   mntr.LogMessage,
		OnChange: mntr.LogMessage,
		OnError:  mntr.LogError,
	}

	return monitor
}

func NewBundle(templator name.Templator) *Bundle {
	monitor := newMonitor()

	bundleConf := &config.Config{
		Monitor:           monitor,
		Orb:               "testsuite",
		CrdName:           "caos_test",
		BaseDirectoryPath: baseDirectoryPath,
		Templator:         templator,
	}

	b := New(bundleConf)
	return b
}

func init() {
	Testmode = true
}

func TestBundle_EmptyApplicationList(t *testing.T) {
	b := NewBundle(yaml.GetName())
	eqApps := b.GetApplications()
	assert.Zero(t, len(eqApps))
}

func TestBundle_AddApplicationsByBundleName(t *testing.T) {
	b := NewBundle(yaml.GetName())

	//Add basisset
	err := b.AddApplicationsByBundleName(bundles.Caos)
	assert.NoError(t, err)
	apps := bundles.GetCaos()

	eqApps := b.GetApplications()
	assert.Equal(t, len(eqApps), len(apps))
	for eqApp := range eqApps {
		assert.Contains(t, apps, eqApp)
	}
}

func TestBundle_AddApplicationsByBundleName_nonexistent(t *testing.T) {
	b := NewBundle(yaml.GetName())
	var nonexistent name.Bundle
	nonexistent = "nonexistent"
	err := b.AddApplicationsByBundleName(nonexistent)
	assert.Error(t, err)
	eqApps := b.GetApplications()
	assert.Equal(t, 0, len(eqApps))
}
func TestBundle_AddApplication(t *testing.T) {
	b := NewBundle(yaml.GetName())

	spec := &v1beta1.ToolsetSpec{}
	app := application.NewTestYAMLApplication(t)

	out, _ := yamlv3.Marshal(testHelperResource)
	app.SetDeploy(spec, true).SetGetYaml(spec, string(out))

	b.AddApplication(app.Application())

	apps := b.GetApplications()
	assert.Equal(t, 1, len(apps))
}

func TestBundle_AddApplication_AlreadyAdded(t *testing.T) {
	b := NewBundle(yaml.GetName())

	spec := &v1beta1.ToolsetSpec{}
	app := application.NewTestYAMLApplication(t)

	out, _ := yamlv3.Marshal(testHelperResource)
	app.SetDeploy(spec, true).SetGetYaml(spec, string(out))

	err := b.AddApplication(app.Application())
	assert.NoError(t, err)

	apps := b.GetApplications()
	assert.Equal(t, 1, len(apps))

	err2 := b.AddApplication(app.Application())
	assert.Error(t, err2)

}

func TestBundle_ReconcileApplication(t *testing.T) {
	b := NewBundle(yaml.GetName())

	spec := &v1beta1.ToolsetSpec{}
	app := application.NewTestYAMLApplication(t)

	out, _ := yamlv3.Marshal(testHelperResource)
	app.SetDeploy(spec, true).SetGetYaml(spec, string(out))

	b.AddApplication(app.Application())

	resources := []*clientgo.Resource{testClientgoResource}

	var wg sync.WaitGroup
	wg.Add(1)
	errChan := make(chan error)
	go b.ReconcileApplication(resources, app.Application().GetName(), spec, &wg, errChan)
	assert.NoError(t, <-errChan)
}

func TestBundle_ReconcileApplication_nonexistent(t *testing.T) {
	b := NewBundle(yaml.GetName())

	spec := &v1beta1.ToolsetSpec{}
	app := application.NewTestYAMLApplication(t)

	out, _ := yamlv3.Marshal(testHelperResource)
	app.SetDeploy(spec, true).SetGetYaml(spec, string(out))

	resources := []*clientgo.Resource{}

	var wg sync.WaitGroup
	wg.Add(1)
	errChan := make(chan error)
	go b.ReconcileApplication(resources, app.Application().GetName(), nil, &wg, errChan)
	assert.Error(t, <-errChan)
}

func TestBundle_Reconcile(t *testing.T) {
	b := NewBundle(yaml.GetName())

	spec := &v1beta1.ToolsetSpec{}
	app := application.NewTestYAMLApplication(t)

	out, _ := yamlv3.Marshal(testHelperResource)
	app.SetDeploy(spec, true).SetGetYaml(spec, string(out))
	b.AddApplication(app.Application())

	resources := []*clientgo.Resource{}

	err := b.Reconcile(resources, spec)
	assert.NoError(t, err)
}

func TestBundle_Reconcile_NoApplications(t *testing.T) {
	b := NewBundle(yaml.GetName())

	spec := &v1beta1.ToolsetSpec{}
	resources := []*clientgo.Resource{}
	err := b.Reconcile(resources, spec)
	assert.NoError(t, err)
}
