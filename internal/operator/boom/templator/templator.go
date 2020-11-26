package templator

import (
	"github.com/caos/orbos/internal/operator/boom/api/latest"
	"github.com/caos/orbos/internal/operator/boom/name"
	"github.com/caos/orbos/internal/operator/boom/templator/helm/chart"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/labels"
)

type Templator interface {
	Template(*labels.API, interface{}, *latest.ToolsetSpec, func(string, string) error) error
	GetResultsFilePath(name.Application, string, string) string
	CleanUp() error
}

type BaseApplication interface {
	GetName() name.Application
}

type YamlApplication interface {
	BaseApplication
	GetYaml(mntr.Monitor, *labels.API, *latest.ToolsetSpec) interface{}
}

type HelmApplication interface {
	BaseApplication
	GetNamespace() string
	SpecToHelmValues(mntr.Monitor, *labels.API, *latest.ToolsetSpec) interface{}
	GetChartInfo() *chart.Chart
}
