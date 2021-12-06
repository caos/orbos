package legacycf

import (
	"context"
	"fmt"
	"time"

	"github.com/caos/orbos/internal/operator/core"
	"github.com/caos/orbos/internal/operator/networking/kinds/networking/legacycf/app"
	"github.com/caos/orbos/internal/operator/networking/kinds/networking/legacycf/config"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/labels"
	"github.com/caos/orbos/pkg/secret/read"
)

func adaptFunc(
	ctx context.Context,
	monitor mntr.Monitor,
	cfg *config.InternalConfig,
) (
	core.QueryFunc,
	core.DestroyFunc,
	core.EnsureFunc,
	error,
) {
	return func(_ kubernetes.ClientInt, _ map[string]interface{}) (core.EnsureFunc, error) {
			return func(k8sClient kubernetes.ClientInt) error {

				groups := make(map[string][]string, 0)
				if cfg.Groups != nil {
					for _, group := range cfg.Groups {
						groups[group.Name] = group.List
					}
				}

				user, err := read.GetSecretValue(k8sClient, cfg.Credentials.User, cfg.Credentials.ExistingUser)
				if err != nil {
					return err
				}
				apiKey, err := read.GetSecretValue(k8sClient, cfg.Credentials.APIKey, cfg.Credentials.ExistingAPIKey)
				if err != nil {
					return err
				}
				userServiceKey, err := read.GetSecretValue(k8sClient, cfg.Credentials.UserServiceKey, cfg.Credentials.ExistingUserServiceKey)
				if err != nil {
					return err
				}

				apps, err := app.New(ctx, cfg.AccountName, user, apiKey, userServiceKey, groups, cfg.Prefix)
				if err != nil {
					return err
				}

				caSecretLabels := labels.MustForName(labels.MustForComponent(cfg.Labels, "cloudflare"), cfg.OriginCASecretName)
				for _, domain := range cfg.Domains {
					err = apps.Ensure(
						ctx,
						cfg.ID,
						k8sClient,
						cfg.Namespace,
						domain.Domain,
						domain.Subdomains,
						domain.Rules,
						caSecretLabels,
						domain.LoadBalancer,
						domain.FloatingIP,
					)
					if err != nil {
						return err
					}
				}
				return nil
			}, nil
		}, func(k8sClient kubernetes.ClientInt) error {
			//TODO
			return nil
		},
		func(k8sClient kubernetes.ClientInt) error {
			monitor.Info("waiting for certificate to be created")
			if err := k8sClient.WaitForSecret(cfg.Namespace, cfg.OriginCASecretName, 60*time.Second); err != nil {
				return fmt.Errorf("error while waiting for certificate secret to be created: %w", err)
			}
			monitor.Info("certificateis created")
			return nil
		}, nil
}
