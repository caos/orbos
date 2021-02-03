package clientgo

import (
	"context"
	"github.com/caos/orbos/mntr"
	pkgerrors "github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetSecret(name, namespace string) (*v1.Secret, error) {
	conf, err := GetClusterConfig(mntr.Monitor{}, "")
	if err != nil {
		return nil, err
	}

	clientset, err := getClientSet(conf)
	if err != nil {
		return nil, err
	}

	secret, err := clientset.CoreV1().Secrets(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	if secret == nil {
		return nil, pkgerrors.New("Secret not found")
	}

	return secret, nil
}
