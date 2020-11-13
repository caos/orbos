package crd

import (
	"os"
	"testing"

	"github.com/caos/orbos/internal/operator/boom/api/migrate"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta1/argocd"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta1/grafana"

	"github.com/caos/orbos/internal/operator/boom/api/v1beta1"
	application "github.com/caos/orbos/internal/operator/boom/application/mock"
	"github.com/caos/orbos/internal/operator/boom/bundle"
	"github.com/caos/orbos/internal/operator/boom/bundle/bundles"
	bundleconfig "github.com/caos/orbos/internal/operator/boom/bundle/config"
	"github.com/caos/orbos/internal/operator/boom/crd/config"
	"github.com/caos/orbos/internal/operator/boom/name"
	"github.com/caos/orbos/internal/operator/boom/templator/yaml"
	"github.com/caos/orbos/internal/utils/clientgo"
	"github.com/caos/orbos/internal/utils/helper"
	"github.com/caos/orbos/mntr"
	"github.com/stretchr/testify/assert"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var (
	fullToolset = &v1beta1.Toolset{
		Metadata: &v1beta1.Metadata{
			Name: "caos_test",
		},
		Spec: &v1beta1.ToolsetSpec{
			Ambassador: &v1beta1.Ambassador{
				Deploy: true,
			},
			Argocd: &argocd.Argocd{
				Deploy: true,
			},
			KubeStateMetrics: &v1beta1.KubeStateMetrics{
				Deploy: true,
			},
			PrometheusOperator: &v1beta1.PrometheusOperator{
				Deploy: true,
			},
			PrometheusNodeExporter: &v1beta1.PrometheusNodeExporter{
				Deploy: true,
			},
			Grafana: &grafana.Grafana{
				Deploy: true,
			},
		},
	}
	changedToolset = &v1beta1.Toolset{
		Metadata: &v1beta1.Metadata{
			Name: "caos_test",
		},
		Spec: &v1beta1.ToolsetSpec{
			Ambassador: &v1beta1.Ambassador{
				Deploy: false,
			},
			Argocd: &argocd.Argocd{
				Deploy: true,
			},
			KubeStateMetrics: &v1beta1.KubeStateMetrics{
				Deploy: true,
			},
			PrometheusOperator: &v1beta1.PrometheusOperator{
				Deploy: true,
			},
			PrometheusNodeExporter: &v1beta1.PrometheusNodeExporter{
				Deploy: true,
			},
			Grafana: &grafana.Grafana{
				Deploy: true,
			},
		},
	}

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

func newCrd() *Crd {

	monitor := mntr.Monitor{
		OnInfo:   mntr.LogMessage,
		OnChange: mntr.LogMessage,
		OnError:  mntr.LogError,
	}

	conf := &config.Config{
		Monitor: monitor,
	}

	return New(conf)
}

func setBundle(c *Crd, bundle name.Bundle) func() {
	basePath := "/tmp/crd_test"
	os.MkdirAll(basePath, os.ModePerm)

	monitor := mntr.Monitor{
		OnInfo:   mntr.LogMessage,
		OnChange: mntr.LogMessage,
		OnError:  mntr.LogError,
	}

	bundleConfig := &bundleconfig.Config{
		Monitor:           monitor,
		CrdName:           "caos_test",
		BundleName:        bundle,
		BaseDirectoryPath: basePath,
		Templator:         yaml.GetName(),
	}

	c.SetBundle(bundleConfig)

	return func() {
		os.RemoveAll(basePath)
	}
}

func init() {
	bundle.Testmode = true
}

func TestNew(t *testing.T) {
	crd := newCrd()
	assert.NotNil(t, crd)
}

func TestNew_noexistendbundle(t *testing.T) {
	var nonexistent name.Bundle
	nonexistent = "nonexistent"
	crd := newCrd()
	clean := setBundle(crd, nonexistent)
	defer clean()
	assert.Error(t, crd.GetStatus())
	assert.NotNil(t, crd)
}

func TestCrd_Reconcile_initial(t *testing.T) {
	crd := newCrd()
	clean := setBundle(crd, bundles.Empty)
	defer clean()
	bundle := crd.GetBundle()

	app := application.NewTestYAMLApplication(t)
	app.SetDeploy(fullToolset.Spec, true).SetGetYaml(fullToolset.Spec, "test")
	bundle.AddApplication(app.Application())
	assert.NotNil(t, crd)

	// when crd is nil
	resources := []*clientgo.Resource{testClientgoResource}
	crd.Reconcile(resources, migrate.V1beta1Tolatest(fullToolset))
	err := crd.GetStatus()
	assert.NoError(t, err)
}

func TestCrd_Reconcile_changed(t *testing.T) {
	crd := newCrd()
	clean := setBundle(crd, bundles.Empty)
	defer clean()
	bundle := crd.GetBundle()

	app := application.NewTestYAMLApplication(t)
	app.SetDeploy(fullToolset.Spec, true).SetGetYaml(fullToolset.Spec, "test")
	bundle.AddApplication(app.Application())
	assert.NotNil(t, crd)

	// when crd is nil
	resources := []*clientgo.Resource{testClientgoResource}
	crd.Reconcile(resources, migrate.V1beta1Tolatest(fullToolset))
	err := crd.GetStatus()
	assert.NoError(t, err)

	//changed crd
	app.SetDeploy(changedToolset.Spec, true).SetGetYaml(changedToolset.Spec, "test2")
	crd.Reconcile(resources, migrate.V1beta1Tolatest(changedToolset))
	err = crd.GetStatus()
	assert.NoError(t, err)
}

func TestCrd_Reconcile_changedDelete(t *testing.T) {
	crd := newCrd()
	clean := setBundle(crd, bundles.Empty)
	defer clean()
	bundle := crd.GetBundle()

	app := application.NewTestYAMLApplication(t)
	app.SetDeploy(fullToolset.Spec, true).SetGetYaml(fullToolset.Spec, "test")
	bundle.AddApplication(app.Application())
	assert.NotNil(t, crd)

	// when crd is nil
	resources := []*clientgo.Resource{testClientgoResource}
	crd.Reconcile(resources, migrate.V1beta1Tolatest(fullToolset))
	err := crd.GetStatus()
	assert.NoError(t, err)

	//changed crd
	app.SetDeploy(changedToolset.Spec, false).SetGetYaml(changedToolset.Spec, "test2")
	crd.Reconcile(resources, migrate.V1beta1Tolatest(changedToolset))
	err = crd.GetStatus()
	assert.NoError(t, err)
}

func TestCrd_Reconcile_initialNotDeployed(t *testing.T) {
	crd := newCrd()
	clean := setBundle(crd, bundles.Empty)
	defer clean()
	bundle := crd.GetBundle()

	app := application.NewTestYAMLApplication(t)
	app.SetDeploy(fullToolset.Spec, false).SetGetYaml(fullToolset.Spec, "test")
	bundle.AddApplication(app.Application())
	assert.NotNil(t, crd)

	// when crd is nil
	resources := []*clientgo.Resource{testClientgoResource}
	crd.Reconcile(resources, migrate.V1beta1Tolatest(fullToolset))
	err := crd.GetStatus()
	assert.NoError(t, err)

	//changed crd
	app.SetDeploy(changedToolset.Spec, false).SetGetYaml(changedToolset.Spec, "test2")
	crd.Reconcile(resources, migrate.V1beta1Tolatest(changedToolset))
	err = crd.GetStatus()
	assert.NoError(t, err)
}
