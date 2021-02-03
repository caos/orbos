package clientgo

import (
	"errors"
	"fmt"
	"github.com/caos/orbos/mntr"
	"os"
	"path/filepath"

	pkgerrors "github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func GetClusterConfig(monitor mntr.Monitor, path string) (*rest.Config, error) {
	if path != "" {
		if cfg, err := getOutClusterConfigPath(path); err == nil {
			return cfg, nil
		}
		monitor.Info(fmt.Sprintf("no kubeconfig under path %s found", path))
	}

	if cfg, err := getOutClusterConfig(); err == nil {
		return cfg, nil
	}
	monitor.Info(fmt.Sprintf("no kubeconfig under path %s found", "$HOME/.kube/config"))

	if cfg, err := getInClusterConfig(); err == nil {
		return cfg, nil
	}
	monitor.Info("no incluster kubeconfig found")
	err := errors.New("no kubeconfig found")
	monitor.Error(err)
	return nil, err
}
func getOutClusterConfigPath(path string) (*rest.Config, error) {
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		return nil, err
	}
	return config, pkgerrors.Wrap(err, "Error while creating out-cluster config")
}

func getOutClusterConfig() (*rest.Config, error) {
	kubeconfig := filepath.Join(homeDir(), ".kube", "config")

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	return config, pkgerrors.Wrap(err, "Error while creating out-cluster config")
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func getInClusterConfig() (*rest.Config, error) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	return config, pkgerrors.Wrap(err, "Error while creating in-cluster config")
}

func getClientSet(config *rest.Config) (*kubernetes.Clientset, error) {
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	return clientset, pkgerrors.Wrap(err, "Error while creating clientset")
}
