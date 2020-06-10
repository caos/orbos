package gitcrd

import (
	"context"
	orbosapi "github.com/caos/orbos/internal/api"
	"github.com/caos/orbos/internal/git"
	"github.com/caos/orbos/internal/operator/boom/api"
	toolsetsv1beta2 "github.com/caos/orbos/internal/operator/boom/api/v1beta2"
	bundleconfig "github.com/caos/orbos/internal/operator/boom/bundle/config"
	"github.com/caos/orbos/internal/operator/boom/cmd"
	"github.com/caos/orbos/internal/operator/boom/crd"
	crdconfig "github.com/caos/orbos/internal/operator/boom/crd/config"
	"github.com/caos/orbos/internal/operator/boom/current"
	"github.com/caos/orbos/internal/operator/boom/gitcrd/config"
	"github.com/caos/orbos/internal/utils/clientgo"
	"github.com/caos/orbos/internal/utils/helper"
	"github.com/caos/orbos/internal/utils/kubectl"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"sync"
)

type GitCrd struct {
	crd              *crd.Crd
	git              *git.Client
	crdDirectoryPath string
	status           error
	monitor          mntr.Monitor
	gitMutex         sync.Mutex
}

func New(conf *config.Config) (*GitCrd, error) {

	monitor := conf.Monitor.WithFields(map[string]interface{}{
		"type": "gitcrd",
	})

	monitor.Info("New GitCRD")

	git := git.New(context.Background(), monitor, conf.User, conf.Email, conf.CrdUrl)
	err := git.Init(conf.PrivateKey)
	if err != nil {
		monitor.Error(err)
		return nil, err
	}

	err = git.Clone()
	if err != nil {
		monitor.Error(err)
		return nil, err
	}

	gitCrd := &GitCrd{
		crdDirectoryPath: conf.CrdDirectoryPath,
		git:              git,
		monitor:          monitor,
		gitMutex:         sync.Mutex{},
	}

	crdConf := &crdconfig.Config{
		Monitor: monitor,
	}

	gitCrd.crd = crd.New(crdConf)

	return gitCrd, nil
}

func (c *GitCrd) GetStatus() error {
	return c.status
}

func (c *GitCrd) SetBackStatus() {
	c.crd.SetBackStatus()
	c.status = nil
}

func (c *GitCrd) SetBundle(conf *bundleconfig.Config) {
	if c.status != nil {
		return
	}

	toolsetCRD, err := c.getCrdMetadata()
	if err != nil {
		c.status = err
		return
	}

	monitor := c.monitor.WithFields(map[string]interface{}{
		"CRD": toolsetCRD.Metadata.Name,
	})

	bundleConf := &bundleconfig.Config{
		Monitor:           monitor,
		CrdName:           toolsetCRD.Metadata.Name,
		BundleName:        conf.BundleName,
		BaseDirectoryPath: conf.BaseDirectoryPath,
		Templator:         conf.Templator,
		Orb:               conf.Orb,
	}

	c.crd.SetBundle(bundleConf)
	c.status = c.crd.GetStatus()
}

func (c *GitCrd) CleanUp() {
	if c.status != nil {
		return
	}

	c.status = os.RemoveAll(c.crdDirectoryPath)
}

func (c *GitCrd) GetRepoURL() string {
	return c.git.GetURL()
}

func (c *GitCrd) Reconcile(currentResourceList []*clientgo.Resource, masterkey string) {
	if c.status != nil {
		return
	}

	monitor := c.monitor.WithFields(map[string]interface{}{
		"action": "reconiling",
	})

	toolsetCRD, err := c.getCrdContent(masterkey)
	if err != nil {
		c.status = err
		return
	}

	if toolsetCRD.Spec.BoomVersion != "" {
		dummyKubeconfig := ""
		if err := cmd.Reconcile(monitor, &dummyKubeconfig, toolsetCRD.Spec.BoomVersion); err != nil {
			c.status = err
			return
		}
	}

	// pre-steps
	if toolsetCRD.Spec.PreApply != nil && toolsetCRD.Spec.PreApply.Deploy == true {
		pre := toolsetCRD.Spec.PreApply
		if pre.Folder == "" {
			c.status = errors.New("PreApply defined but no folder provided")
			return
		}

		folderExists, err := c.git.ExistsFolder(pre.Folder)
		if err != nil {
			c.status = err
			return
		}
		if !folderExists {
			c.status = errors.New("PreApply provided folder is nonexistent")
			return
		}

		folderEmpty, err := c.git.EmptyFolder(pre.Folder)
		if err != nil {
			c.status = err
			return
		}
		if folderEmpty {
			c.status = errors.New("PreApply provided folder is empty")
			return
		}

		c.gitMutex.Lock()
		err = helper.CopyFolderToLocal(c.git, c.crdDirectoryPath, pre.Folder)
		c.gitMutex.Unlock()
		if err != nil {
			c.status = err
			return
		}

		if err := useFolder(monitor, pre.Deploy, c.crdDirectoryPath, pre.Folder); err != nil {
			c.status = err
			return
		}

	}

	c.crd.Reconcile(currentResourceList, toolsetCRD)
	err = c.crd.GetStatus()
	if err != nil {
		c.status = err
		return
	}

	// post-steps
	if toolsetCRD.Spec.PostApply != nil && toolsetCRD.Spec.PostApply.Deploy == true {
		post := toolsetCRD.Spec.PostApply
		if post.Folder == "" {
			c.status = errors.New("PostApply defined but no folder provided")
		}

		folderExists, err := c.git.ExistsFolder(post.Folder)
		if err != nil {
			c.status = err
			return
		}
		if !folderExists {
			c.status = errors.New("PostApply provided folder is nonexistent")
			return
		}

		folderEmpty, err := c.git.EmptyFolder(post.Folder)
		if err != nil {
			c.status = err
			return
		}
		if folderEmpty {
			c.status = errors.New("PostApply provided folder is empty")
			return
		}

		c.gitMutex.Lock()
		err = helper.CopyFolderToLocal(c.git, c.crdDirectoryPath, post.Folder)
		c.gitMutex.Unlock()
		if err != nil {
			c.status = err
			return
		}

		if err := useFolder(monitor, post.Deploy, c.crdDirectoryPath, post.Folder); err != nil {
			c.status = err
			return
		}
	}
}

func (c *GitCrd) getCrdMetadata() (*toolsetsv1beta2.ToolsetMetadata, error) {
	c.gitMutex.Lock()
	defer c.gitMutex.Unlock()

	if err := c.git.Clone(); err != nil {
		return nil, err
	}

	toolsetCRD := &toolsetsv1beta2.ToolsetMetadata{}
	err := c.git.ReadYamlIntoStruct("boom.yml", toolsetCRD)
	if err != nil {
		return nil, errors.Wrap(err, "Error while unmarshaling boom.yml to struct")
	}

	return toolsetCRD, nil

}

func (c *GitCrd) getCrdContent(masterkey string) (*toolsetsv1beta2.Toolset, error) {
	c.gitMutex.Lock()
	defer c.gitMutex.Unlock()

	desiredTree, err := orbosapi.ReadBoomYml(c.git)
	if err != nil {
		return nil, err
	}

	desiredKind, _, err := api.ParseToolset(desiredTree, masterkey)
	if err != nil {
		return nil, errors.Wrap(err, "parsing desired state failed")
	}
	desiredTree.Parsed = desiredKind

	return desiredKind, nil
}

func (c *GitCrd) WriteBackCurrentState(currentResourceList []*clientgo.Resource, masterkey string) {
	if c.status != nil {
		return
	}

	content, err := yaml.Marshal(current.ResourcesToYaml(currentResourceList))
	if err != nil {
		c.status = err
		return
	}

	toolsetCRD, err := c.getCrdContent(masterkey)
	if err != nil {
		c.status = err
		return
	}

	currentFolder := toolsetCRD.Spec.CurrentStateFolder
	if currentFolder == "" {
		currentFolder = filepath.Join("caos-internal", "boom")
	}

	file := git.File{
		Path:    filepath.Join(currentFolder, "current.yaml"),
		Content: content,
	}

	c.gitMutex.Lock()
	defer c.gitMutex.Unlock()
	c.status = c.git.UpdateRemote("current state changed", file)
}

func useFolder(monitor mntr.Monitor, deploy bool, tempDirectory, folderRelativePath string) error {
	folderPath := filepath.Join(tempDirectory, folderRelativePath)

	command := kubectl.NewApply(folderPath).Build()
	if !deploy {
		command = kubectl.NewDelete(folderPath).Build()
	}

	return helper.Run(monitor, command)
}
