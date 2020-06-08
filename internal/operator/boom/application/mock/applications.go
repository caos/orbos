package application

import (
	"testing"

	"github.com/caos/orbos/internal/operator/boom/api/v1beta1"
	"github.com/caos/orbos/internal/operator/boom/name"
	"github.com/golang/mock/gomock"
)

const (
	Name      name.Application = "mockapplication"
	Namespace                  = "mocknamespace"
)

type TestApplication struct {
	app *MockApplication
}

func NewTestApplication(t *testing.T) *TestApplication {

	mockApp := NewMockApplication(gomock.NewController(t))
	mockApp.EXPECT().GetName().AnyTimes().DoAndReturn(Name)
	app := &TestApplication{
		app: mockApp,
	}

	return app
}

func (a *TestApplication) Application() *MockApplication {
	return a.app
}

func (a *TestApplication) SetDeploy(spec *v1beta1.ToolsetSpec, b bool) *TestApplication {
	a.app.EXPECT().Deploy(spec).AnyTimes().Return(b)
	return a
}

type TestYAMLApplication struct {
	app *MockYAMLApplication
}

func NewTestYAMLApplication(t *testing.T) *TestYAMLApplication {

	mockApp := NewMockYAMLApplication(gomock.NewController(t))
	mockApp.EXPECT().GetName().AnyTimes().Return(Name)
	app := &TestYAMLApplication{
		app: mockApp,
	}

	return app
}

func (a *TestYAMLApplication) Application() *MockYAMLApplication {
	return a.app
}

func (a *TestYAMLApplication) SetDeploy(spec *v1beta1.ToolsetSpec, b bool) *TestYAMLApplication {
	a.app.EXPECT().Deploy(spec).AnyTimes().Return(b)
	return a
}

func (a *TestYAMLApplication) SetGetYaml(spec *v1beta1.ToolsetSpec, yaml string) *TestYAMLApplication {
	a.app.EXPECT().GetYaml(gomock.Any(), spec).AnyTimes().Return(yaml)
	return a
}
