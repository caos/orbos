package core

import (
	"crypto/rsa"
	"github.com/caos/orbos/internal/operator/core"
	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/tree"
)

type CurrentDBList struct {
	Common  *tree.Common `yaml:",inline"`
	Current *DatabaseCurrentDBList
}

type DatabaseCurrentDBList struct {
	Databases []string
}

func (c *CurrentDBList) GetURL() string {
	return ""
}

func (c *CurrentDBList) GetPort() string {
	return ""
}

func (c *CurrentDBList) GetReadyQuery() core.EnsureFunc {
	return nil
}

func (c *CurrentDBList) GetCertificateKey() *rsa.PrivateKey {
	return nil
}

func (c *CurrentDBList) SetCertificateKey(key *rsa.PrivateKey) {
	return
}

func (c *CurrentDBList) GetCertificate() []byte {
	return nil
}

func (c *CurrentDBList) SetCertificate(cert []byte) {
	return
}

func (c *CurrentDBList) GetListDatabasesFunc() func(k8sClient *kubernetes.Client) ([]string, error) {
	return func(k8sClient *kubernetes.Client) ([]string, error) {
		return c.Current.Databases, nil
	}
}

func (c *CurrentDBList) GetListUsersFunc() func(k8sClient *kubernetes.Client) ([]string, error) {
	return nil
}

func (c *CurrentDBList) GetAddUserFunc() func(user string) (core.QueryFunc, error) {
	return nil
}

func (c *CurrentDBList) GetDeleteUserFunc() func(user string) (core.DestroyFunc, error) {
	return nil
}
