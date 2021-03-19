package cli

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/caos/orbos/internal/api"
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
) (k8sClient *kubernetes.Client, fromOrbiter bool, err error) {

	orbConfigIsIncompleteErr := orb.IsComplete(orbConfig)
	if orbConfigIsIncompleteErr != nil && gitops {
		return nil, fromOrbiter, orbConfigIsIncompleteErr
	}

	var kc string
	if orbConfigIsIncompleteErr == nil && gitops {
		if err := gitClient.Configure(orbConfig.URL, []byte(orbConfig.Repokey)); err != nil {
			return k8sClient, fromOrbiter, err
		}

		if err := gitClient.Clone(); err != nil {
			return k8sClient, fromOrbiter, err
		}
		var err error
		fromOrbiter, err = api.ExistsOrbiterYml(gitClient)
		if err != nil {
			return k8sClient, fromOrbiter, err
		}

		if fromOrbiter {
			orbTree, err := api.ReadOrbiterYml(gitClient)
			if err != nil {
				return k8sClient, fromOrbiter, errors.New("failed to parse orbiter.yml")
			}

			orbDef, err := orb2.ParseDesiredV0(orbTree)
			if err != nil {
				return k8sClient, fromOrbiter, errors.New("failed to parse orbiter.yml")
			}

			for clustername, _ := range orbDef.Clusters {
				path := strings.Join([]string{"orbiter", clustername, "kubeconfig", "encrypted"}, ".")

				kc, err = secret.Read(
					nil,
					path,
					operators.GetAllSecretsFunc(monitor, gitops, gitClient, nil, orbConfig),
				)
				if err != nil || kc == "" {
					if kc == "" && err == nil {
						err = errors.New("no kubeconfig found")
					}
					return nil, fromOrbiter, fmt.Errorf("failed to get kubeconfig: %w", err)
				}
			}
		}
	}

	if kc == "" {
		value, err := ioutil.ReadFile(kubeconfig)
		if err != nil {
			return k8sClient, fromOrbiter, err
		}
		kc = string(value)
	}

	k8sClient = kubernetes.NewK8sClient(monitor, &kc)
	if !k8sClient.Available() {
		return nil, fromOrbiter, errors.New("Kubernetes is not connectable")
	}

	return k8sClient, fromOrbiter, nil
}
