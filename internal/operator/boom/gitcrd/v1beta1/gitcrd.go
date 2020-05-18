package v1beta1

import (
	"github.com/caos/orbos/internal/operator/boom/api"
	"github.com/caos/orbos/internal/tree"
	helper2 "github.com/caos/orbos/internal/utils/helper"
	"os"
	"path/filepath"
	"sync"

	toolsetsv1beta1 "github.com/caos/orbos/internal/operator/boom/api/v1beta1"
	bundleconfig "github.com/caos/orbos/internal/operator/boom/bundle/config"
	"github.com/caos/orbos/internal/operator/boom/crd"
	"github.com/caos/orbos/internal/operator/boom/crd/v1beta1"
	"github.com/caos/orbos/internal/operator/boom/gitcrd/v1beta1/config"
	"github.com/caos/orbos/internal/utils/clientgo"
	"github.com/caos/orbos/internal/utils/helper"
	"github.com/caos/orbos/internal/utils/kubectl"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/caos/orbos/internal/git"
	crdconfig "github.com/caos/orbos/internal/operator/boom/crd/config"
	"github.com/caos/orbos/internal/operator/boom/current"
)

type GitCrd struct {
	crd              crd.Crd
	git              *git.Client
	crdDirectoryPath string
	crdPath          string
	status           error
	monitor          mntr.Monitor
	gitMutex         sync.Mutex
}

func New(conf *config.Config) (*GitCrd, error) {

	monitor := conf.Monitor.WithFields(map[string]interface{}{
		"version": "v1beta1",
	})

	gitConf := *conf.Git
	gitCrd := &GitCrd{
		crdDirectoryPath: conf.CrdDirectoryPath,
		crdPath:          conf.CrdPath,
		git:              &gitConf,
		monitor:          monitor,
		gitMutex:         sync.Mutex{},
	}

	crdConf := &crdconfig.Config{
		Monitor: monitor,
		Version: v1beta1.GetVersion(),
	}

	crd, err := crd.New(crdConf)
	if err != nil {
		return nil, err
	}

	gitCrd.crd = crd

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
func (c *GitCrd) GetRepoCRDPath() string {
	return c.crdPath
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

	// pre-steps
	if toolsetCRD.Spec.PreApply != nil && toolsetCRD.Spec.PreApply.Deploy == true {
		pre := toolsetCRD.Spec.PreApply
		if pre.Folder == "" {
			c.status = errors.New("PreApply defined but no folder provided")
		}
		if !helper2.FolderExists(pre.Folder) {
			c.status = errors.New("PreApply provided folder is nonexistent")
		}
		if empty, err := helper2.FolderEmpty(pre.Folder); empty == true || err != nil {
			c.status = errors.New("PreApply provided folder is empty")
		}

		c.gitMutex.Lock()
		err := helper2.CopyFolderToLocal(c.git, c.crdDirectoryPath, pre.Folder)
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
		if !helper2.FolderExists(post.Folder) {
			c.status = errors.New("PostApply provided folder is nonexistent")
		}
		if empty, err := helper2.FolderEmpty(post.Folder); empty == true || err != nil {
			c.status = errors.New("PostApply provided folder is empty")
		}

		c.gitMutex.Lock()
		err := helper2.CopyFolderToLocal(c.git, c.crdDirectoryPath, post.Folder)
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

func (c *GitCrd) getCrdMetadata() (*toolsetsv1beta1.ToolsetMetadata, error) {
	c.gitMutex.Lock()
	defer c.gitMutex.Unlock()

	if err := c.git.Clone(); err != nil {
		return nil, err
	}

	toolsetCRD := &toolsetsv1beta1.ToolsetMetadata{}
	err := c.git.ReadYamlIntoStruct(c.crdPath, toolsetCRD)
	if err != nil {
		return nil, errors.Wrapf(err, "Error while unmarshaling yaml %s to struct", c.crdPath)
	}

	return toolsetCRD, nil

}

func (c *GitCrd) getCrdContent(masterkey string) (*toolsetsv1beta1.Toolset, error) {
	c.gitMutex.Lock()
	defer c.gitMutex.Unlock()

	if err := c.git.Clone(); err != nil {
		return nil, err
	}

	raw := c.git.Read(c.crdPath)

	desiredTree := &tree.Tree{}
	if err := yaml.Unmarshal(raw, desiredTree); err != nil {
		return nil, err
	}

	desiredKind, err := api.ParseToolset(desiredTree, masterkey)
	if err != nil {
		return nil, errors.Wrap(err, "parsing desired state failed")
	}
	desiredTree.Parsed = desiredKind

	/*err := c.git.ReadYamlIntoStruct(c.crdPath, toolsetCRD)
	if err != nil {
		return nil, errors.Wrapf(err, "Error while unmarshaling yaml %s to struct", c.crdPath)
	}*/

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
