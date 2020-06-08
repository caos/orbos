package application

//go:generate mockgen -source application.go -package application -destination mock/application.mock.go github.com/caos/internal/bundle/application HelmApplication,YAMLApplication
