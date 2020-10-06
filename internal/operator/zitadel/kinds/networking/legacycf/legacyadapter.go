package legacycf

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/zitadel"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/networking/legacycf/app"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/networking/legacycf/config"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
)

func adaptFunc(
	monitor mntr.Monitor,
	cfg *config.InternalConfig,
) (
	zitadel.QueryFunc,
	zitadel.DestroyFunc,
	zitadel.EnsureFunc,
	error,
) {
	return func(_ *kubernetes.Client, _ map[string]interface{}) (zitadel.EnsureFunc, error) {
			return func(k8sClient *kubernetes.Client) error {

				internalLabels := map[string]string{}
				for k, v := range cfg.Labels {
					internalLabels[k] = v
				}
				internalLabels["app.kubernetes.io/component"] = "networking"

				groups := make(map[string][]string, 0)
				if cfg.Groups != nil {
					for _, group := range cfg.Groups {
						groups[group.Name] = group.List
					}
				}

				apps, err := app.New(cfg.Credentials.User.Value, cfg.Credentials.APIKey.Value, cfg.Credentials.UserServiceKey.Value, groups, cfg.Prefix)
				if err != nil {
					return err
				}

				for _, domain := range cfg.Domains {
					err = apps.Ensure(k8sClient, cfg.Namespace, internalLabels, domain.Domain, domain.Subdomains, domain.Rules, cfg.OriginCASecretName)
					if err != nil {
						return err
					}
				}
				return nil
			}, nil
		}, func(k8sClient *kubernetes.Client) error {
			//TODO
			return nil
		},
		func(k8sClient *kubernetes.Client) error {
			monitor.Info("waiting for certificate to be created")
			if err := k8sClient.WaitForSecret(cfg.Namespace, cfg.OriginCASecretName, 60); err != nil {
				return errors.Wrap(err, "error while waiting for certificate secret to be created")
			}
			monitor.Info("certificateis created")
			return nil
		}, nil
}
