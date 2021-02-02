package kubernetes

//go:generate mockgen -source client.go -package kubernetesmock -destination mock/client.mock.go github.com/caos/pkg/kubernetes ClientInt
//go:generate goderive .
