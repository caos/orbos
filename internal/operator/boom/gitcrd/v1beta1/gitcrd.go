package v1beta1

import (
	"github.com/caos/orbos/internal/operator/boom/api"
	"github.com/caos/orbos/internal/tree"
	helper2 "github.com/caos/orbos/internal/utils/helper"
	"github.com/caos/orbos/internal/utils/kustomize"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
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
	if toolsetCRD.Spec.PreApply != nil {
		preapplymonitor := monitor.WithField("application", "preapply")
		preapplymonitor.Info("Start")
		if err := c.applyFolder(preapplymonitor, toolsetCRD.Spec.PreApply, toolsetCRD.Spec.ForceApply); err != nil {
			c.status = errors.Wrap(err, "Preapply failed")
			return
		}
		preapplymonitor.Info("Done")
	}

	c.crd.Reconcile(currentResourceList, toolsetCRD)
	err = c.crd.GetStatus()
	if err != nil {
		c.status = err
		return
	}

	// post-steps
	if toolsetCRD.Spec.PostApply != nil {
		preapplymonitor := monitor.WithField("application", "postapply")
		preapplymonitor.Info("Start")
		if err := c.applyFolder(monitor, toolsetCRD.Spec.PostApply, toolsetCRD.Spec.ForceApply); err != nil {
			c.status = errors.Wrap(err, "Postapply failed")
			return
		}
		preapplymonitor.Info("Done")
	}
}

func (c *GitCrd) applyFolder(monitor mntr.Monitor, apply *toolsetsv1beta1.Apply, force bool) error {
	if apply.Folder == "" {
		return errors.New("No folder provided")
	}

	c.gitMutex.Lock()
	err := helper2.CopyFolderToLocal(c.git, c.crdDirectoryPath, apply.Folder)
	c.gitMutex.Unlock()
	if err != nil {
		return err
	}

	localFolder := filepath.Join(c.crdDirectoryPath, apply.Folder)
	if !helper2.FolderExists(localFolder) {
		return errors.New("Folder is nonexistent")
	}

	if empty, err := helper2.FolderEmpty(localFolder); empty == true || err != nil {
		return errors.New("Provided folder is empty")
	}

	if err := useFolder(monitor, apply.Deploy, localFolder, force); err != nil {
		return err
	}
	return nil
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

	desiredKind, err := api.ParseToolset(desiredTree)
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

func useFolder(monitor mntr.Monitor, deploy bool, folderPath string, force bool) error {

	files, err := getFilesInDirectory(folderPath)
	if err != nil {
		return err
	}
	kustomizeFile := false
	for _, file := range files {
		if strings.HasSuffix(file, "kustomization.yaml") {
			kustomizeFile = true
			break
		}
	}

	if kustomizeFile {
		command, err := kustomize.New(folderPath)
		if err != nil {
			return err
		}
		command = command.Apply(force)
		if !deploy {
			command = command.Delete()
		}
		return helper.Run(monitor, command.Build())
	} else {
		return recursiveFolder(monitor, folderPath, deploy, force)
	}
}

func recursiveFolder(monitor mntr.Monitor, folderPath string, deploy, force bool) error {
	command := kubectl.NewApply(folderPath).Build()
	if force {
		command = kubectl.NewApply(folderPath).Force().Build()
	}
	if !deploy {
		command = kubectl.NewDelete(folderPath).Build()
	}

	folders, err := getDirsInDirectory(folderPath)
	if err != nil {
		return err
	}

	for _, folder := range folders {
		if folderPath != folder {
			if err := recursiveFolder(monitor, filepath.Join(folderPath, folder), deploy, force); err != nil {
				return err
			}
		}
	}
	return helper.Run(monitor, command)
}

func getFilesInDirectory(dirPath string) ([]string, error) {
	files := make([]string, 0)

	infos, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	for _, info := range infos {
		if !info.IsDir() {
			files = append(files, filepath.Join(dirPath, info.Name()))
		}
	}

	return files, err
}

func getDirsInDirectory(dirPath string) ([]string, error) {
	dirs := make([]string, 0)

	infos, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	for _, info := range infos {
		if info.IsDir() {
			dirs = append(dirs, filepath.Join(dirPath, info.Name()))
		}
	}
	return dirs, err
}
