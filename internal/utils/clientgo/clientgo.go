package clientgo

import (
	"os"
	"path/filepath"

	pkgerrors "github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var InConfig = true

func getClusterConfig() (*rest.Config, error) {
	if InConfig {
		return getInClusterConfig()
	} else {
		return getOutClusterConfig()
	}
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
