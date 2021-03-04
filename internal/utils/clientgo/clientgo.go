package clientgo

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/caos/orbos/mntr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func GetClusterConfig(monitor mntr.Monitor, path string) (*rest.Config, error) {

	monitor.Debug("trying to use in-cluster kube client")
	if cfg, err := getInClusterConfig(); err == nil {
		monitor.Debug("using in-cluster kube client")
		return cfg, nil
	}

	if path != "" {
		if cfg, err := getOutClusterConfigPath(path); err == nil {
			return cfg, nil
		}
		monitor.Info(fmt.Sprintf("no kubeconfig found at path %s", path))
	}

	if cfg, err := getOutClusterConfig(); err == nil {
		return cfg, nil
	}

	monitor.Info(fmt.Sprintf("no kubeconfig found at path %s", "$HOME/.kube/config"))

	err := errors.New("no kubeconfig found")
	monitor.Error(err)
	return nil, err
}
func getOutClusterConfigPath(path string) (*rest.Config, error) {
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", path)
	return config, nonNilErrorf("creating out-cluster config failed: %w", err)
}

func getOutClusterConfig() (*rest.Config, error) {
	kubeconfig := filepath.Join(homeDir(), ".kube", "config")

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	return config, nonNilErrorf("creating out-cluster config failed: %w", err)
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
	return config, nonNilErrorf("creating in-cluster config failed: %w", err)
}

func getClientSet(config *rest.Config) (*kubernetes.Clientset, error) {
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	return clientset, nonNilErrorf("creating clientset failed: %w", err)
}

func nonNilErrorf(format string, err error, val ...interface{}) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf(format, append([]interface{}{err}, val...))
}
