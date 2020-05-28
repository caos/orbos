package orbgit

import (
	"context"
	"errors"
	"github.com/caos/orbos/internal/git"
	orbconfig "github.com/caos/orbos/internal/orb"
	"github.com/caos/orbos/internal/ssh"
	"github.com/caos/orbos/internal/stores/github"
	"github.com/caos/orbos/mntr"
	"math/rand"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	Comitter  string
	Email     string
	OrbConfig *orbconfig.Orb
	Action    string
}

func NewGitClient(ctx context.Context, monitor mntr.Monitor, conf *Config) (*git.Client, func(), error) {
	deployKeyPriv := ""
	deployKeyDelete := func() {}

	if conf.OrbConfig.Repokey == "" {
		dir := filepath.Dir(conf.OrbConfig.Path)

		deployKeyPrivLocal, deployKeyPub, err := ssh.Generate()
		if err != nil {
			return nil, deployKeyDelete, errors.New("failed to generate ssh key for deploy key")
		}
		g := github.New(monitor).LoginOAuth(dir)
		if g.GetStatus() != nil {
			return nil, deployKeyDelete, errors.New("failed github oauth login ")
		}
		repo, err := g.GetRepositorySSH(conf.OrbConfig.URL)
		if err != nil {
			return nil, deployKeyDelete, errors.New("failed to get github repository")
		}

		desc := strings.Join([]string{"orbos", conf.Action, generateRandom()}, "-")

		if err := g.CreateDeployKey(repo, desc, deployKeyPub).GetStatus(); err != nil {
			return nil, deployKeyDelete, errors.New("failed to create deploy keys in repository")
		}
		deployKeyPriv = deployKeyPrivLocal

		deployKeyDelete = func() {
			if err := g.DeleteDeployKeysByDescription(repo, desc).GetStatus(); err != nil {
				monitor.Error(errors.New("failed to clear deploy keys in repository"))
			}
		}

	} else {
		deployKeyPriv = conf.OrbConfig.Repokey
	}

	gitClient := git.New(ctx, monitor, conf.Comitter, conf.Email, conf.OrbConfig.URL)
	if err := gitClient.Init([]byte(deployKeyPriv)); err != nil {
		panic(err)
	}

	return gitClient, deployKeyDelete, nil
}

func generateRandom() string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789")
	length := 8
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	return b.String()
}
