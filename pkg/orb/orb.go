package orb

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/caos/orbos/internal/helpers"
	"github.com/caos/orbos/internal/ssh"
	"github.com/caos/orbos/internal/stores/github"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/git"
	"github.com/caos/orbos/pkg/secret"
)

var alphanum = regexp.MustCompile("[^a-zA-Z0-9]+")

type Orb struct {
	id        string `yaml:"-"`
	Path      string `yaml:"-"`
	URL       string
	Repokey   string
	Masterkey string
}

func (o *Orb) IsConnectable() (err error) {
	defer func() {
		if err != nil {
			err = mntr.ToUserError(fmt.Errorf("repository is not connectable: %w", err))
		}
	}()
	if o.URL == "" {
		err = helpers.Concat(err, errors.New("repository url is missing"))
	}

	if o.Repokey == "" {
		err = helpers.Concat(err, errors.New("repository key is missing"))
	}
	return err
}

func IsComplete(o *Orb) (err error) {

	defer func() {
		if err != nil {
			err = mntr.ToUserError(fmt.Errorf("orbconfig is incomplete: %w", err))
		}
	}()

	if o == nil {
		return errors.New("path not provided")
	}

	if o.Masterkey == "" {
		err = helpers.Concat(err, errors.New("master key is missing"))
	}

	if o.Path == "" {
		err = helpers.Concat(err, errors.New("file path is missing"))
	}

	return helpers.Concat(err, o.IsConnectable())
}

func ParseOrbConfig(orbConfigPath string) (orb *Orb, err error) {

	defer func() {
		if err != nil {
			err = mntr.ToUserError(fmt.Errorf("parsing orbconfig failed: %w", err))
		}
	}()

	gitOrbConfig, err := ioutil.ReadFile(orbConfigPath)

	if err != nil {
		return nil, fmt.Errorf("unable to read orbconfig: %w", err)
	}

	orb = &Orb{}
	if err := yaml.Unmarshal(gitOrbConfig, orb); err != nil {
		return nil, fmt.Errorf("unable to unmarshal yaml: %w", err)
	}

	orb.Path = orbConfigPath
	secret.Masterkey = orb.Masterkey
	return orb, nil
}

func (o *Orb) writeBackOrbConfig() error {

	data, err := yaml.Marshal(o)
	if err != nil {
		panic(err)
	}

	if err := ioutil.WriteFile(o.Path, data, os.ModePerm); err != nil {
		return mntr.ToUserError(fmt.Errorf("writing orbconfig failed: %w", err))
	}
	return nil
}

func Reconfigure(
	ctx context.Context,
	monitor mntr.Monitor,
	orbConfig *Orb,
	newRepoURL,
	newMasterKey,
	newRepoKey string,
	gitClient *git.Client,
	clientID,
	clientSecret string) (err error) {

	defer func() {
		if err != nil {
			err = fmt.Errorf("reconfiguring orb failed: %w", err)
		}
	}()

	if orbConfig.URL == "" && newRepoURL == "" {
		return mntr.ToUserError(errors.New("repository url is neighter passed by flag repourl nor written in orbconfig"))
	}

	// TODO: Remove?
	if orbConfig.URL != "" && newRepoURL != "" && orbConfig.URL != newRepoURL {
		return mntr.ToUserError(fmt.Errorf("repository url %s is not reconfigurable", orbConfig.URL))
	}

	if orbConfig.Masterkey == "" && newMasterKey == "" {
		return mntr.ToUserError(errors.New("master key is neighter passed by flag masterkey nor written in orbconfig"))
	}

	var changes bool
	if newMasterKey != "" {
		monitor.Info("Changing masterkey in current orbconfig")
		if orbConfig.Masterkey == "" {
			secret.Masterkey = newMasterKey
		}
		orbConfig.Masterkey = newMasterKey
		changes = true
	}
	if newRepoURL != "" {
		monitor.Info("Changing repository url in current orbconfig")
		defer func() {
			if err == nil {
				monitor.WithField("url", newRepoURL).CaptureMessage("New Repository URL configured")
			}
		}()
		orbConfig.URL = newRepoURL
		changes = true
	}

	if newRepoKey != "" {
		monitor.Info("Changing used key to connect to repository in current orbconfig")
		orbConfig.Repokey = newRepoKey
		changes = true
	}

	configureGit := func(mustConfigure bool) error {
		if err := gitClient.Configure(orbConfig.URL, []byte(orbConfig.Repokey)); err != nil {
			if mustConfigure {
				panic(err)
			}
			return err
		}
		if err := gitClient.Clone(); err != nil {
			// this is considered a user error, therefore no panic
			return err
		}
		// this is considered a user error, therefore no panic
		return gitClient.Check()
	}

	// If the repokey already has read/write permissions, don't generate a new one.
	// This ensures git providers other than github keep being supported
	// Only if you're not trying to set a new key, as you don't want to generate a new key then
	if err := configureGit(false); err != nil && newRepoKey == "" {

		monitor.Info("Starting connection with git-repository")

		dir := filepath.Dir(orbConfig.Path)

		deployKeyPrivLocal, deployKeyPub := ssh.Generate()
		g := github.New(monitor).LoginOAuth(ctx, dir, clientID, clientSecret)
		if err := g.GetStatus(); err != nil {
			return fmt.Errorf("github oauth login failed: %w", err)
		}
		repo, err := g.GetRepositorySSH(orbConfig.URL)
		if err != nil {
			return fmt.Errorf("failed to get github repository: %w", err)
		}

		if err := g.EnsureNoDeployKey(repo).GetStatus(); err != nil {
			return fmt.Errorf("failed to clear deploy keys in repository: %w", err)
		}

		if err := g.CreateDeployKey(repo, deployKeyPub).GetStatus(); err != nil {
			return fmt.Errorf("failed to create deploy keys in repository: %w", err)
		}
		orbConfig.Repokey = deployKeyPrivLocal

		if err := configureGit(true); err != nil {
			return err
		}
		changes = true
	}

	if changes {
		monitor.Info("Writing local orbconfig")
		if err := orbConfig.writeBackOrbConfig(); err != nil {
			return err
		}
	}

	return nil
}

func (o *Orb) ID() (id string, err error) {

	defer func() {
		err = mntr.ToUserError(err)
	}()

	if err := IsComplete(o); err != nil {
		return "", err
	}

	if o.id != "" {
		return o.id, nil
	}

	o.id = alphanum.ReplaceAllString(strings.TrimSuffix(strings.TrimPrefix(o.URL, "git@"), ".git"), "-")
	return o.id, nil
}
