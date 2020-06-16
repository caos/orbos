package gitcrd

import (
	"context"
	"github.com/caos/orbos/internal/operator/boom/metrics"
	"github.com/caos/orbos/internal/utils/random"
	"strings"

	"github.com/caos/orbos/internal/git"
	bundleconfig "github.com/caos/orbos/internal/operator/boom/bundle/config"
	"github.com/caos/orbos/internal/operator/boom/gitcrd/config"
	"github.com/caos/orbos/internal/operator/boom/gitcrd/v1beta1"
	v1beta1config "github.com/caos/orbos/internal/operator/boom/gitcrd/v1beta1/config"
	"github.com/caos/orbos/internal/utils/clientgo"
	"github.com/caos/orbos/internal/utils/helper"
	"github.com/pkg/errors"
)

type GitCrd interface {
	SetBundle(*bundleconfig.Config)
	Reconcile([]*clientgo.Resource, string)
	WriteBackCurrentState([]*clientgo.Resource, string)
	CleanUp()
	GetStatus() error
	SetBackStatus()
	GetRepoURL() string
	GetRepoCRDPath() string
}

func New(conf *config.Config) (GitCrd, error) {

	conf.Monitor.Info("New GitCRD")

	gitClient := git.New(context.Background(), conf.Monitor, conf.User, conf.Email, conf.CrdUrl)
	err := gitClient.Init(conf.PrivateKey)
	if err != nil {
		conf.Monitor.Error(err)
		return nil, err
	}

	if err := gitClient.ReadCheck(); err != nil {
		conf.Monitor.Error(err)
		return nil, err
	}

	if err := gitClient.WriteCheck(random.Generate()); err != nil {
		conf.Monitor.Error(err)
		return nil, err
	}

	err = gitClient.Clone()
	if err != nil {
		conf.Monitor.Error(err)
		metrics.FailedGitClone(conf.CrdUrl)
		return nil, err
	}
	metrics.SuccessfulGitClone(conf.CrdUrl)

	crdFileStruct := &helper.Resource{}
	if err := gitClient.ReadYamlIntoStruct(conf.CrdPath, crdFileStruct); err != nil {
		conf.Monitor.Error(err)
		return nil, err
	}

	group := "boom.caos.ch"
	version := "v1beta1"

	parts := strings.Split(crdFileStruct.ApiVersion, "/")
	if parts[0] != group {
		err := errors.Errorf("Unknown CRD apiGroup %s", parts[0])
		conf.Monitor.Error(err)
		return nil, err
	}

	if parts[1] != version {
		err := errors.Errorf("Unknown CRD version %s", parts[1])
		conf.Monitor.Error(err)
		return nil, err
	}

	monitor := conf.Monitor.WithFields(map[string]interface{}{
		"type": "gitcrd",
	})

	v1beta1conf := &v1beta1config.Config{
		Monitor:          monitor,
		Git:              gitClient,
		CrdDirectoryPath: conf.CrdDirectoryPath,
		CrdPath:          conf.CrdPath,
	}

	return v1beta1.New(v1beta1conf)
}
