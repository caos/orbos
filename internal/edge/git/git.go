package git

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/caos/orbiter/logging"
	"golang.org/x/crypto/ssh"
	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/memfs"
	gogit "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	gitssh "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	"gopkg.in/src-d/go-git.v4/storage/memory"
	"gopkg.in/yaml.v2"
)

type Client struct {
	logger    logging.Logger
	ctx       context.Context
	committer string
	auth      *gitssh.PublicKeys
	repo      *gogit.Repository
	fs        billy.Filesystem
	workTree  *gogit.Worktree
	progress  io.Writer
	repoURL   string
}

func New(ctx context.Context, logger logging.Logger, committer string, repoURL string) *Client {
	newClient := &Client{
		ctx:       ctx,
		logger:    logger,
		committer: committer,
		repoURL:   repoURL,
	}

	if logger.IsVerbose() {
		newClient.progress = os.Stdout
	}
	return newClient
}

func (g *Client) Init(deploykey []byte) error {
	signer, err := ssh.ParsePrivateKey(deploykey)
	if err != nil {
		return errors.Wrap(err, "parsing deployment key failed")
	}

	g.auth = &gitssh.PublicKeys{
		User:   "git",
		Signer: signer,
	}

	// TODO: Fix
	g.auth.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	return nil
}

func (g *Client) Clone() error {

	g.fs = memfs.New()

	var err error
	g.repo, err = gogit.CloneContext(g.ctx, memory.NewStorage(), g.fs, &gogit.CloneOptions{
		URL:          g.repoURL,
		Auth:         g.auth,
		SingleBranch: true,
		Depth:        1,
		Progress:     g.progress,
	})
	if err != nil {
		return errors.Wrapf(err, "cloning repository from %s failed", g.repoURL)
	}
	g.logger.Debug("Repository cloned")

	g.workTree, err = g.repo.Worktree()
	if err != nil {
		return errors.Wrapf(err, "getting worktree from repository with url %s failed", g.repoURL)
	}

	return nil
}

/*
func (g *Client) Pull() error {
	g.logger.Debug("Pulling")

	err := g.workTree.PullContext(g.ctx, &gogit.PullOptions{
		//			Depth:        1,
		SingleBranch: true,
		RemoteName:   "origin",
		Auth:         g.auth,
		Progress:     g.progress,
		Force:        true,
	})
	if err != nil && !strings.Contains(err.Error(), gogit.NoErrAlreadyUpToDate.Error()) {
		return errors.Wrap(err, "pulling repository failed")
	}

	g.logger.Debug("Repository pulled to worktree")
	return nil
}
*/
func (g *Client) Read(path string) (map[string]interface{}, error) {
	readLogger := g.logger.WithFields(map[string]interface{}{
		"path": path,
	})
	readLogger.Debug("Reading file")
	file, err := g.fs.Open(path)
	if err != nil {
		if os.IsNotExist(errors.Cause(err)) {
			return make(map[string]interface{}), nil
		}
		return nil, errors.Wrapf(err, "opening %s from worktree failed", path)
	}
	defer file.Close()
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, errors.Wrapf(err, "reading %s from worktree failed", path)
	}
	if readLogger.IsVerbose() {
		readLogger.Debug("File read")
		fmt.Println(string(fileBytes))
	}
	unmarshalled := make(map[string]interface{})
	if err := yaml.Unmarshal(fileBytes, unmarshalled); err != nil {
		fmt.Println(string(fileBytes))
		return nil, errors.Wrapf(err, "unmarshalling %s from worktree failed", path)
	}
	readLogger.Debug("File parsed")
	return unmarshalled, nil
}

type File struct {
	Path      string
	Overwrite func(map[string]interface{}) ([]byte, error)
	Force     bool
}

type LatestFile struct {
	Path    string
	Content []byte
}

func (g *Client) UpdateRemoteUntilItWorks(file *File) ([]byte, error) {

	if err := g.Clone(); err != nil {
		return nil, errors.Wrap(err, "recloning before committing changes failed")
	}

	newContent, err := g.Read(file.Path)
	if err != nil && !file.Force {
		return nil, errors.Wrap(err, "reloading file before committing changes failed")
	}

	overwritten, err := file.Overwrite(newContent)
	if err != nil {
		return nil, err
	}

	if err := g.updateAndStage(file.Path, overwritten); err != nil {
		return nil, err
	}

	status, err := g.workTree.Status()
	if err != nil {
		return nil, errors.Wrap(err, "querying worktree status failed")
	}

	if status.IsClean() {
		g.logger.Info("No changes")
		return overwritten, nil
	}

	if err := g.commit(); err != nil {
		return nil, err
	}

	if err := g.push(); err != nil && strings.Contains(err.Error(), "command error on refs/heads/master: cannot lock ref 'refs/heads/master': is at ") {
		g.logger.Debug("Undoing latest commit")
		if resetErr := g.workTree.Reset(&gogit.ResetOptions{
			Mode: gogit.HardReset,
		}); resetErr != nil {
			return overwritten, errors.Wrap(resetErr, "undoing the latest commit failed")
		}

		newLatestFiles, err := g.UpdateRemoteUntilItWorks(file)
		return newLatestFiles, errors.Wrap(err, "pushing failed")
	}
	return overwritten, nil
}

func (g *Client) updateAndStage(path string, content []byte) error {
	updateLogger := g.logger.WithFields(map[string]interface{}{
		"path": path,
	})

	updateLogger.Debug("Overwriting local index")

	file, err := g.fs.Create(path)
	if err != nil {
		return errors.Wrapf(err, "creating file %s in worktree failed", path)
	}
	defer file.Close()

	if _, err := io.Copy(file, bytes.NewReader(content)); err != nil {
		return errors.Wrapf(err, "writing file %s in worktree failed", path)
	}

	_, err = g.workTree.Add(path)
	if err != nil {
		updateLogger.Debug("Changes staged")
	}
	return errors.Wrapf(err, "staging worktree changes in file %s failed", path)
}

func (g *Client) push() error {

	err := g.repo.PushContext(g.ctx, &gogit.PushOptions{
		RemoteName: "origin",
		//			RefSpecs:   refspecs,
		Auth:     g.auth,
		Progress: g.progress,
	})
	if err != nil {
		return errors.Wrap(err, "pushing repository failed")
	}

	g.logger.Info("Repository pushed")
	return nil
}

func (g *Client) commit() error {
	if _, err := g.workTree.Commit("update current state or secrets", &gogit.CommitOptions{
		Author: &object.Signature{
			Name:  g.committer,
			Email: "hi@caos.ch",
			When:  time.Now(),
		},
	}); err != nil {
		return errors.Wrap(err, "committing changes failed")
	}
	g.logger.Debug("Changes commited")
	return nil
}
