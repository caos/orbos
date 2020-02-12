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
)

type Client struct {
	logger    logging.Logger
	ctx       context.Context
	committer string
	email     string
	auth      *gitssh.PublicKeys
	repo      *gogit.Repository
	fs        billy.Filesystem
	workTree  *gogit.Worktree
	progress  io.Writer
	repoURL   string
}

func New(ctx context.Context, logger logging.Logger, committer, email, repoURL string) *Client {
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

func (g *Client) Read(path string) ([]byte, error) {
	readLogger := g.logger.WithFields(map[string]interface{}{
		"path": path,
	})
	readLogger.Debug("Reading file")
	file, err := g.fs.Open(path)
	if err != nil {
		if os.IsNotExist(errors.Cause(err)) {
			return make([]byte, 0), nil
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
	return fileBytes, nil
}

type File struct {
	Path    string
	Content []byte
}

func (g *Client) Commit(msg string, files ...File) (bool, error) {
	clean, err := g.stage(files...)
	if err != nil {
		return false, err
	}

	if clean {
		return false, nil
	}

	return true, g.commit(msg)
}

func (g *Client) UpdateRemote(msg string, files ...File) error {
	if err := g.Clone(); err != nil {
		return errors.Wrap(err, "recloning before committing changes failed")
	}

	changed, err := g.Commit(msg, files...)
	if err != nil {
		return err
	}

	if !changed {
		g.logger.Info(false, "No changes")
		return nil
	}

	return g.Push()
}

func (g *Client) UpdateRemoteUntilItWorks(msg string, path string, overwrite func([]byte) ([]byte, error), force bool) ([]byte, error) {

	if err := g.Clone(); err != nil {
		return nil, errors.Wrap(err, "recloning before committing changes failed")
	}

	newContent, err := g.Read(path)
	if err != nil && !force {
		return nil, errors.Wrap(err, "reloading file before committing changes failed")
	}

	overwritten, err := overwrite(newContent)
	if err != nil {
		return nil, err
	}

	clean, err := g.stage(File{Path: path, Content: overwritten})
	if err != nil {
		return nil, err
	}

	if clean {
		g.logger.Info(false, "No changes")
		return overwritten, nil
	}

	if err := g.commit(msg); err != nil {
		return nil, err
	}

	if err := g.Push(); err != nil && strings.Contains(err.Error(), "command error on refs/heads/master: cannot lock ref 'refs/heads/master': is at ") {
		g.logger.Debug("Undoing latest commit")
		if resetErr := g.workTree.Reset(&gogit.ResetOptions{
			Mode: gogit.HardReset,
		}); resetErr != nil {
			return overwritten, errors.Wrap(resetErr, "undoing the latest commit failed")
		}

		newLatestFiles, err := g.UpdateRemoteUntilItWorks(msg, path, overwrite, force)
		return newLatestFiles, errors.Wrap(err, "pushing failed")
	}
	return overwritten, nil
}

func (g *Client) stage(files ...File) (bool, error) {
	for _, f := range files {
		updateLogger := g.logger.WithFields(map[string]interface{}{
			"path": f.Path,
		})

		updateLogger.Debug("Overwriting local index")

		file, err := g.fs.Create(f.Path)
		if err != nil {
			return true, errors.Wrapf(err, "creating file %s in worktree failed", f.Path)
		}
		defer file.Close()

		if _, err := io.Copy(file, bytes.NewReader(f.Content)); err != nil {
			return true, errors.Wrapf(err, "writing file %s in worktree failed", f.Path)
		}

		_, err = g.workTree.Add(f.Path)
		if err != nil {
			return true, errors.Wrapf(err, "staging worktree changes in file %s failed", f.Path)
		}
	}

	status, err := g.workTree.Status()
	if err != nil {
		return true, errors.Wrap(err, "querying worktree status failed")
	}

	return status.IsClean(), nil
}

func (g *Client) commit(msg string) error {

	if _, err := g.workTree.Commit(msg, &gogit.CommitOptions{
		Author: &object.Signature{
			Name:  g.committer,
			Email: g.email,
			When:  time.Now(),
		},
	}); err != nil {
		return errors.Wrap(err, "committing changes failed")
	}
	g.logger.Debug("Changes commited")
	return nil
}

func (g *Client) Push() error {

	err := g.repo.PushContext(g.ctx, &gogit.PushOptions{
		RemoteName: "origin",
		//			RefSpecs:   refspecs,
		Auth:     g.auth,
		Progress: g.progress,
	})
	if err != nil {
		return errors.Wrap(err, "pushing repository failed")
	}

	g.logger.Info(false, "Repository pushed")
	return nil
}
