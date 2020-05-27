package github

import (
	"bufio"
	"context"
	"fmt"
	"github.com/caos/oidc/pkg/cli"
	"github.com/caos/oidc/pkg/rp"
	"github.com/caos/orbos/mntr"
	"github.com/google/go-github/v31/github"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/oauth2"
	githubOAuth "golang.org/x/oauth2/github"
	"os"
	"strings"
	"syscall"
)

var (
	ClientID     = ""
	ClientSecret = ""
	Key          = ""
)

type githubAPI struct {
	monitor mntr.Monitor
	client  *github.Client
	status  error
}

func (g *githubAPI) GetStatus() error {
	return g.status
}

func New(monitor mntr.Monitor) *githubAPI {
	githubMonitor := monitor.WithFields(map[string]interface{}{
		"store": "github",
	})
	return &githubAPI{
		client:  nil,
		status:  nil,
		monitor: githubMonitor,
	}
}

func (g *githubAPI) IsLoggedIn() bool {
	return g.client != nil
}

func (g *githubAPI) Login() *githubAPI {
	r := bufio.NewReader(os.Stdin)
	fmt.Print("GitHub Username: ")
	username, _ := r.ReadString('\n')

	fmt.Print("GitHub Password: ")
	bytePassword, _ := terminal.ReadPassword(int(syscall.Stdin))
	password := string(bytePassword)

	g.LoginBasicAuth(username, password)

	// Is this a two-factor auth error? If so, prompt for OTP and try again.
	if _, ok := g.status.(*github.TwoFactorAuthError); ok {
		g.status = nil

		fmt.Print("\nGitHub OTP: ")
		otp, _ := r.ReadString('\n')

		g.LoginTwoFactor(username, password, otp)
		if g.GetStatus() != nil {
			return g
		}
	} else if g.status != nil {
		g.client = nil
	}

	return g
}

func (g *githubAPI) LoginOAuth() *githubAPI {
	port := "9999"
	callbackPath := "/orbctl/github/callback"

	rpConfig := &rp.Config{
		ClientID:     ClientID,
		ClientSecret: ClientSecret,
		CallbackURL:  fmt.Sprintf("http://localhost:%v%v", port, callbackPath),
		Scopes:       []string{"repo", "repo_deployment"},
		Endpoints:    githubOAuth.Endpoint,
	}

	oauth2Client := cli.CodeFlowForClient(rpConfig, []byte(Key), callbackPath, port)

	g.monitor.Info("Finished CodeFlow")

	client := github.NewClient(oauth2Client)

	ctx := context.Background()
	_, _, g.status = client.Users.Get(ctx, "")
	if g.GetStatus() != nil {
		return g
	}

	g.monitor.Info("Successful CodeFlow")

	g.client = client
	return g
}

func (g *githubAPI) LoginToken(token string) *githubAPI {
	if g.status != nil {
		return g
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	_, _, g.status = client.Users.Get(ctx, "")
	if g.GetStatus() != nil {
		return g
	}

	g.monitor.Info("Successful PersonalAccessTokenFlow")
	g.client = client
	return g
}

func (g *githubAPI) LoginBasicAuth(username, password string) *githubAPI {
	if g.status != nil {
		return g
	}

	tp := github.BasicAuthTransport{
		Username: strings.TrimSpace(username),
		Password: strings.TrimSpace(password),
	}

	client := github.NewClient(tp.Client())

	ctx := context.Background()
	_, _, g.status = client.Users.Get(ctx, "")
	if g.GetStatus() != nil {
		return g
	}

	g.monitor.Info("Successful BasicAuthFlow")
	g.client = client
	return g
}

func (g *githubAPI) LoginTwoFactor(username, password, twoFactor string) *githubAPI {
	if g.status != nil {
		return g
	}

	tp := github.BasicAuthTransport{
		Username: strings.TrimSpace(username),
		Password: strings.TrimSpace(password),
		OTP:      strings.TrimSpace(twoFactor),
	}

	client := github.NewClient(tp.Client())

	ctx := context.Background()
	_, _, g.status = client.Users.Get(ctx, "")
	if g.GetStatus() != nil {
		return g
	}

	g.monitor.Info("Successful BasicAuthFlow with OTP")
	g.client = client
	return g
}

func (g *githubAPI) GetRepositorySSH(url string) (*github.Repository, error) {
	if g.GetStatus() != nil {
		return nil, g.status
	}

	ctx := context.Background()
	parts := strings.Split(strings.TrimPrefix(url, "git@github.com:"), "/")

	repo, _, err := g.client.Repositories.Get(ctx, parts[0], strings.TrimSuffix(parts[1], ".git"))
	if err != nil {
		g.status = err
	}
	return repo, err
}

func (g *githubAPI) GetRepositories() ([]*github.Repository, error) {
	if g.GetStatus() != nil {
		return nil, g.status
	}

	ctx := context.Background()
	repos := make([]*github.Repository, 0)
	addRepos, err := addRepositories(ctx, g.client, "private", "owner")
	if err != nil {
		g.status = err
		return nil, err
	}
	repos = append(repos, addRepos...)

	addRepos, err = addRepositories(ctx, g.client, "public", "owner")
	if err != nil {
		g.status = err
		return nil, err
	}
	repos = append(repos, addRepos...)

	addRepos, err = addRepositories(ctx, g.client, "private", "organization_member")
	if err != nil {
		g.status = err
		return nil, err
	}
	repos = append(repos, addRepos...)

	addRepos, err = addRepositories(ctx, g.client, "public", "organization_member")
	if err != nil {
		g.status = err
		return nil, err
	}
	repos = append(repos, addRepos...)

	addRepos, err = addRepositories(ctx, g.client, "private", "collaborator")
	if err != nil {
		g.status = err
		return nil, err
	}
	repos = append(repos, addRepos...)

	addRepos, err = addRepositories(ctx, g.client, "public", "collaborator")
	if err != nil {
		g.status = err
		return nil, err
	}
	repos = append(repos, addRepos...)

	return repos, nil
}

func addRepositories(ctx context.Context, client *github.Client, visibility, affiliation string) ([]*github.Repository, error) {
	opts := &github.RepositoryListOptions{
		Visibility:  visibility,
		Affiliation: affiliation,
	}

	addRepos, _, err := client.Repositories.List(ctx, "", opts)
	return addRepos, err
}

func (g *githubAPI) GetDeployKeys(repo *github.Repository) []*github.Key {
	if g.GetStatus() != nil {
		return nil
	}

	ctx := context.Background()

	keys, _, err := g.client.Repositories.ListKeys(ctx, *repo.Owner.Login, *repo.Name, nil)
	if err != nil {
		g.status = err
		return nil
	}
	return keys
}

func (g *githubAPI) CreateDeployKey(repo *github.Repository, description string, value string) *githubAPI {
	if g.GetStatus() != nil {
		return g
	}
	ctx := context.Background()

	f := false
	key := github.Key{
		Key:      &value,
		Title:    &description,
		ReadOnly: &f,
	}

	_, _, g.status = g.client.Repositories.CreateKey(ctx, *repo.Owner.Login, *repo.Name, &key)

	return g
}

func (g *githubAPI) DeleteDeployKeysByDescription(repo *github.Repository, description string) *githubAPI {
	if g.GetStatus() != nil {
		return g
	}
	ctx := context.Background()
	keys := g.GetDeployKeys(repo)
	if g.status != nil {
		return g
	}

	for _, key := range keys {
		if *key.Title == description {
			if _, g.status = g.client.Repositories.DeleteKey(ctx, *repo.Owner.Login, *repo.Name, *key.ID); g.status != nil {
				return g
			}
		}
	}

	return g
}
