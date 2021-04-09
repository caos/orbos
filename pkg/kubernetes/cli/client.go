package cli

import (
	"fmt"
	"io/ioutil"
	"strings"

	orb2 "github.com/caos/orbos/internal/operator/orbiter/kinds/orb"
	"github.com/caos/orbos/internal/secret/operators"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/git"
	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/orb"
	"github.com/caos/orbos/pkg/secret"
	"github.com/pkg/errors"
)

func Client(
	monitor mntr.Monitor,
	orbConfig *orb.Orb,
	gitClient *git.Client,
	kubeconfig string,
	gitops bool,
) (*kubernetes.Client, error) {

	var kc string
	orbConfigIsIncompleteErr := orb.IsComplete(orbConfig)
	if orbConfigIsIncompleteErr != nil && gitops {
		return nil, orbConfigIsIncompleteErr
	}

	if orbConfigIsIncompleteErr == nil && gitops {
		if err := gitClient.Configure(orbConfig.URL, []byte(orbConfig.Repokey)); err != nil {
			return nil, err
		}

		if err := gitClient.Clone(); err != nil {
			return nil, err
		}
		if gitClient.Exists(git.OrbiterFile) {
			orbTree, err := gitClient.ReadTree(git.OrbiterFile)
			if err != nil {
				return nil, errors.New("failed to parse orbiter.yml")
			}

			orbDef, err := orb2.ParseDesiredV0(orbTree)
			if err != nil {
				return nil, errors.New("failed to parse orbiter.yml")
			}

			for clustername, _ := range orbDef.Clusters {
				path := strings.Join([]string{"orbiter", clustername, "kubeconfig", "encrypted"}, ".")

				kc, err = secret.Read(
					nil,
					path,
					operators.GetAllSecretsFunc(monitor, false, gitops, gitClient, nil, orbConfig),
				)
				if err != nil || kc == "" {
					if kc == "" && err == nil {
						err = errors.New("no kubeconfig found")
					}
					return nil, fmt.Errorf("failed to get kubeconfig: %w", err)
				}
			}
		}
	}

	if kc == "" {
		value, err := ioutil.ReadFile(kubeconfig)
		if err != nil {
			return nil, err
		}
		kc = string(value)
	}

	return kubernetes.NewK8sClient(monitor, &kc)
}
