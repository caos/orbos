package types

import (
	"github.com/caos/orbos/internal/operator/boom/api/latest"
	"github.com/caos/orbos/internal/operator/boom/name"
	"github.com/caos/orbos/internal/operator/boom/templator/helm/chart"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/labels"
)

type Application interface {
	Deploy(*latest.ToolsetSpec) bool
	GetName() name.Application
}

type HelmApplication interface {
	Application
	GetNamespace() string
	GetChartInfo() *chart.Chart
	GetImageTags() map[string]string
	SpecToHelmValues(mntr.Monitor, *labels.API, *latest.ToolsetSpec) interface{}
}

type YAMLApplication interface {
	Application
	GetYaml(mntr.Monitor, *labels.API, *latest.ToolsetSpec) interface{}
}
