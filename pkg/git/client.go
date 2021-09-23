package git

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"errors"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v3"

	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/tree"
)

type DesiredFile string

func (d DesiredFile) WOExtension() string {
	return strings.Split(string(d), ".")[0]
}

const (
	writeCheckTag = "writecheck"

	OrbiterFile    DesiredFile = "orbiter.yml"
	BoomFile       DesiredFile = "boom.yml"
	NetworkingFile DesiredFile = "networking.yml"
	DatabaseFile   DesiredFile = "database.yml"
	ZitadelFile    DesiredFile = "zitadel.yml"
)

type Client struct {
	monitor   mntr.Monitor
	ctx       context.Context
	committer string
	email     string
	auth      *gitssh.PublicKeys
	repo      *gogit.Repository
	fs        billy.Filesystem
	storage   *memory.Storage
	workTree  *gogit.Worktree
	progress  io.Writer
	repoURL   string
	cloned    bool
}

func New(ctx context.Context, monitor mntr.Monitor, committer, email string) *Client {
	newClient := &Client{
		ctx:       ctx,
		committer: committer,
		email:     email,
		monitor:   monitor,
		storage:   memory.NewStorage(),
		fs:        memfs.New(),
	}

	if monitor.IsVerbose() {
		newClient.progress = os.Stdout
	}
	return newClient
}

func (g *Client) GetURL() string {
	return g.repoURL
}

func (g *Client) Configure(repoURL string, deploykey []byte) error {
	signer, err := ssh.ParsePrivateKey(deploykey)
	if err != nil {
		return fmt.Errorf("parsing deployment key failed: %w", err)
	}

	if repoURL != g.repoURL {
		g.repoURL = repoURL
		g.cloned = false
	}
	g.monitor = g.monitor.WithField("repository", repoURL)

	g.auth = &gitssh.PublicKeys{
		User:   "git",
		Signer: signer,
	}

	// TODO: Fix
	g.auth.HostKeyCallback = ssh.InsecureIgnoreHostKey()

	return nil
}

func (g *Client) Check() error {
	if !g.cloned {
		return nil
	}
	if err := g.readCheck(); err != nil {
		return err
	}

	return g.writeCheck()
}

func (g *Client) readCheck() error {

	rem := gogit.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{g.repoURL},
	})

	// We can then use every Remote functions to retrieve wanted information
	_, err := rem.List(&gogit.ListOptions{
		Auth: g.auth,
	})
	if err != nil {
		return mntr.ToUserError(fmt.Errorf("read check failed: %w", err))
	}

	g.monitor.Info("Read check success")
	return nil
}

func (g *Client) writeCheck() (err error) {

	defer func() {
		err = mntr.ToUserError(err)
	}()

	head, err := g.repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get head: %w", err)
	}
	localWriteCheckTag := strings.Join([]string{writeCheckTag, g.committer}, "-")

	ref, createErr := g.repo.CreateTag(localWriteCheckTag, head.Hash(), nil)
	if createErr == gogit.ErrTagExists {
		if ref, err = g.repo.Tag(localWriteCheckTag); err != nil {
			return err
		}
		createErr = nil
	}

	if createErr != nil {
		return fmt.Errorf("write-check failed: %w", createErr)
	}

	if pushErr := g.repo.Push(&gogit.PushOptions{
		RemoteName: "origin",
		RefSpecs: []config.RefSpec{
			config.RefSpec("+" + ref.Name() + ":" + ref.Name()),
		},
		Auth: g.auth,
	}); pushErr != nil && pushErr != gogit.NoErrAlreadyUpToDate {
		return fmt.Errorf("write-check failed: %w", pushErr)
	}

	g.monitor.Debug("Write check tag created")

	if deleteErr := g.repo.DeleteTag(localWriteCheckTag); deleteErr != nil && deleteErr != gogit.ErrTagNotFound {
		return fmt.Errorf("write-check cleanup delete tag failed: %w", deleteErr)
	}

	if err := g.repo.Push(&gogit.PushOptions{
		RemoteName: "origin",
		RefSpecs: []config.RefSpec{
			config.RefSpec(":" + ref.Name()),
		},
		Auth: g.auth,
	}); err != nil {
		return fmt.Errorf("write-check cleanup failed: %w", err)
	}

	g.monitor.Debug("Write check tag cleaned up")
	g.monitor.Info("Write check success")
	return nil
}

func (g *Client) Clone() (err error) {
	for i := 0; i < 10; i++ {
		if err = g.clone(); err == nil {
			return nil
		}
		time.Sleep(time.Second)
	}
	return err
}

func (g *Client) clone() error {
	g.fs = memfs.New()

	g.monitor.Debug("Cloning")
	var err error
	g.repo, err = gogit.CloneContext(g.ctx, memory.NewStorage(), g.fs, &gogit.CloneOptions{
		URL:          g.repoURL,
		Auth:         g.auth,
		SingleBranch: true,
		Depth:        1,
		Progress:     g.progress,
	})
	if err != nil {
		return mntr.ToUserError(fmt.Errorf("cloning repository from %s failed: %w", g.repoURL, err))
	}
	g.monitor.Debug("Cloned")

	g.workTree, err = g.repo.Worktree()
	if err != nil {
		panic(err)
	}

	g.cloned = true

	return nil
}

func (g *Client) Read(path string) []byte {

	readmonitor := g.monitor.WithFields(map[string]interface{}{
		"path": path,
	})
	readmonitor.Debug("Reading file")
	file, err := g.fs.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make([]byte, 0)
		}
		panic(err)
	}
	defer file.Close()
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}
	if readmonitor.IsVerbose() {
		readmonitor.Debug("File read")
		fmt.Println(string(fileBytes))
	}
	return fileBytes
}

func (g *Client) ReadYamlIntoStruct(path string, struc interface{}) error {
	data := g.Read(path)

	err := yaml.Unmarshal(data, struc)
	if err != nil {
		err = fmt.Errorf("unmarshaling yaml %s to struct failed: %w", path, err)
	}

	return err
}

func (g *Client) ExistsFolder(path string) (bool, error) {
	monitor := g.monitor.WithFields(map[string]interface{}{
		"path": path,
	})
	monitor.Debug("Reading folder")
	_, err := g.fs.ReadDir(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("opening %s from worktree failed: %w", path, err)
	}

	return true, nil
}

func (g *Client) EmptyFolder(path string) (bool, error) {
	monitor := g.monitor.WithFields(map[string]interface{}{
		"path": path,
	})
	monitor.Debug("Reading folder")
	files, err := g.fs.ReadDir(path)
	if err != nil {
		return false, fmt.Errorf("opening %s from worktree failed: %w", path, err)
	}
	if len(files) == 0 {
		return true, nil
	}
	return false, nil
}

func (g *Client) ReadFolder(path string) (map[string][]byte, []string, error) {
	monitor := g.monitor.WithFields(map[string]interface{}{
		"path": path,
	})
	monitor.Debug("Reading folder")
	dirBytes := make(map[string][]byte, 0)
	files, err := g.fs.ReadDir(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string][]byte, 0), nil, nil
		}
		return nil, nil, fmt.Errorf("opening %s from worktree failed: %w", path, err)
	}
	subdirs := make([]string, 0)
	for _, file := range files {
		if !file.IsDir() {
			filePath := filepath.Join(path, file.Name())
			fileBytes := g.Read(filePath)
			dirBytes[file.Name()] = fileBytes
		} else {
			subdirs = append(subdirs, file.Name())
		}
	}

	if monitor.IsVerbose() {
		monitor.Debug("Folder read")
		fmt.Println(dirBytes)
	}
	return dirBytes, subdirs, nil
}

type File struct {
	Path    string
	Content []byte
}

func (g *Client) stageAndCommit(msg string, files ...File) (bool, error) {
	if g.stage(files...) {
		return false, nil
	}

	return true, g.Commit(msg)
}

func (g *Client) UpdateRemote(msg string, whenCloned func() []File) error {

	if err := g.Clone(); err != nil {
		return fmt.Errorf("recloning before committing changes failed: %w", err)
	}

	changed, err := g.stageAndCommit(msg, whenCloned()...)
	if err != nil {
		return err
	}

	if !changed {
		g.monitor.Info("No changes")
		return nil
	}
	err = g.Push()
	if err != nil &&
		(errors.Is(err, plumbing.ErrObjectNotFound) ||
			strings.Contains(err.Error(), "cannot lock ref")) {
		g.monitor.WithField("response", err.Error()).Info("Git collision detected, retrying")
		return g.UpdateRemote(msg, whenCloned)
	}
	return err
}

func (g *Client) stage(files ...File) bool {
	for _, f := range files {
		updatemonitor := g.monitor.WithFields(map[string]interface{}{
			"path": f.Path,
		})

		updatemonitor.Debug("Overwriting local index")

		file, err := g.fs.Create(f.Path)
		if err != nil {
			panic(err)
		}
		//noinspection GoDeferInLoop
		defer file.Close()

		if _, err := io.Copy(file, bytes.NewReader(f.Content)); err != nil {
			panic(err)
		}

		_, err = g.workTree.Add(f.Path)
		if err != nil {
			panic(err)
		}
	}

	status, err := g.workTree.Status()
	if err != nil {
		panic(err)
	}

	return status.IsClean()
}

func (g *Client) Commit(msg string) error {

	if _, err := g.workTree.Commit(msg, &gogit.CommitOptions{
		Author: &object.Signature{
			Name:  g.committer,
			Email: g.email,
			When:  time.Now(),
		},
	}); err != nil {
		return fmt.Errorf("committing changes failed: %w", err)
	}
	g.monitor.Debug("Changes commited")
	return nil
}

func (g *Client) Push() error {

	if err := g.repo.PushContext(g.ctx, &gogit.PushOptions{
		RemoteName: "origin",
		//			RefSpecs:   refspecs,
		Auth:     g.auth,
		Progress: g.progress,
	}); err != nil {
		return fmt.Errorf("pushing repository failed: %w", err)
	}

	g.monitor.Info("Repository pushed")
	return nil
}

func (g *Client) Exists(path DesiredFile) bool {
	of := g.Read(string(path))
	if of != nil && len(of) > 0 {
		return true
	}
	return false
}

func (g *Client) ReadTree(path DesiredFile) (*tree.Tree, error) {
	tree := &tree.Tree{}
	return tree, yaml.Unmarshal(g.Read(string(path)), tree)
}

type GitDesiredState struct {
	Desired *tree.Tree
	Path    DesiredFile
}

func (g *Client) PushGitDesiredStates(monitor mntr.Monitor, msg string, desireds []GitDesiredState) (err error) {
	monitor.OnChange = func(_ string, fields map[string]string) {
		err = g.UpdateRemote(mntr.SprintCommit(msg, fields), func() []File {
			gitFiles := make([]File, len(desireds))
			for i := range desireds {
				desired := desireds[i]
				gitFiles[i] = File{
					Path:    string(desired.Path),
					Content: common.MarshalYAML(desired.Desired),
				}
			}
			return gitFiles
		})
	}
	monitor.Changed(msg)
	return err
}

func (g *Client) PushDesiredFunc(file DesiredFile, desired *tree.Tree) func(mntr.Monitor) error {
	return func(monitor mntr.Monitor) error {
		monitor.WithField("file", file).Info("Writing desired state")
		return g.PushGitDesiredStates(monitor, fmt.Sprintf("Desired state written to %s", file), []GitDesiredState{{
			Desired: desired,
			Path:    file,
		}})
	}
}
