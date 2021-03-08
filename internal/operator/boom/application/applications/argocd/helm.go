package argocd

import (
	"github.com/caos/orbos/internal/utils/helper"
	"strings"

	toolsetslatest "github.com/caos/orbos/internal/operator/boom/api/latest"
	"github.com/caos/orbos/internal/operator/boom/application/applications/argocd/config/credential"
	"github.com/caos/orbos/internal/operator/boom/application/applications/argocd/config/repository"

	"github.com/caos/orbos/internal/operator/boom/application/applications/argocd/config"
	"github.com/caos/orbos/internal/operator/boom/application/applications/argocd/customimage"
	"github.com/caos/orbos/internal/operator/boom/application/applications/argocd/helm"
	"github.com/caos/orbos/internal/operator/boom/templator/helm/chart"
	"github.com/caos/orbos/mntr"
)

func (a *Argocd) HelmPreApplySteps(monitor mntr.Monitor, toolsetCRDSpec *toolsetslatest.ToolsetSpec) ([]interface{}, error) {
	secrets := make([]interface{}, 0)
	if toolsetCRDSpec.Reconciling == nil {
		return secrets, nil
	}
	customimagesecrets := customimage.GetSecrets(toolsetCRDSpec.Reconciling)
	repoSecrets := repository.GetSecrets(toolsetCRDSpec.Reconciling)
	credSecrets := credential.GetSecrets(toolsetCRDSpec.Reconciling)

	secrets = append(secrets, customimagesecrets...)
	secrets = append(secrets, repoSecrets...)
	secrets = append(secrets, credSecrets...)
	return secrets, nil
}

func (a *Argocd) HelmMutate(monitor mntr.Monitor, toolsetCRDSpec *toolsetslatest.ToolsetSpec, resultFilePath string) error {
	if toolsetCRDSpec.Reconciling != nil && toolsetCRDSpec.Reconciling.CustomImage != nil && toolsetCRDSpec.Reconciling.CustomImage.Enabled {
		spec := toolsetCRDSpec.Reconciling

		if spec.CustomImage.GopassStores != nil && len(spec.CustomImage.GopassStores) > 0 {
			if err := customimage.AddPostStartFromSpec(spec, resultFilePath); err != nil {
				return err
			}
		}
	}

	return nil
}

func (a *Argocd) SpecToHelmValues(monitor mntr.Monitor, toolsetCRDSpec *toolsetslatest.ToolsetSpec) interface{} {
	imageTags := a.GetImageTags()
	image := "argoproj/argocd"

	if toolsetCRDSpec != nil && toolsetCRDSpec.Reconciling != nil {
		helper.OverwriteExistingValues(imageTags, map[string]string{
			image: toolsetCRDSpec.Reconciling.OverwriteVersion,
		})
		helper.OverwriteExistingKey(imageTags, &image, toolsetCRDSpec.Reconciling.OverwriteImage)
	}
	values := helm.DefaultValues(imageTags, image)

	spec := toolsetCRDSpec.Reconciling
	if spec == nil {
		return values
	}

	if spec.CustomImage != nil && spec.CustomImage.Enabled {
		conf := customimage.FromSpec(spec, imageTags)
		values.RepoServer.Image = &helm.Image{
			Repository:      conf.ImageRepository,
			Tag:             conf.ImageTag,
			ImagePullPolicy: "IfNotPresent",
		}
		if conf.AddSecretVolumes != nil {
			for _, v := range conf.AddSecretVolumes {
				items := make([]*helm.Item, 0)
				for _, item := range v.Secret.Items {
					items = append(items, &helm.Item{Key: item.Key, Path: item.Path})
				}

				values.RepoServer.Volumes = append(values.RepoServer.Volumes, &helm.Volume{
					Secret: &helm.VolumeSecret{
						SecretName:  v.Secret.SecretName,
						Items:       items,
						DefaultMode: v.DefaultMode,
					},
					Name: v.Name,
				})
			}
		}
		if conf.AddVolumeMounts != nil {
			for _, v := range conf.AddVolumeMounts {
				values.RepoServer.VolumeMounts = append(values.RepoServer.VolumeMounts, &helm.VolumeMount{
					Name:      v.Name,
					MountPath: v.MountPath,
					SubPath:   v.SubPath,
					ReadOnly:  v.ReadOnly,
				})
			}
		}
	}

	conf := config.GetFromSpec(monitor, spec)
	if conf.Repositories != "" && conf.Repositories != "[]\n" {
		values.Server.Config.Repositories = conf.Repositories
	}
	if conf.Credentials != "" && conf.Credentials != "[]\n" {
		values.Server.Config.RepositoryCredentials = conf.Credentials
	}

	if conf.ConfigManagementPlugins != "" {
		values.Server.Config.ConfigManagementPlugins = conf.ConfigManagementPlugins
	}

	if spec.Network != nil && spec.Network.Domain != "" {

		if conf.OIDC != "" {
			values.Server.Config.OIDC = conf.OIDC
		}

		if conf.Connectors != "" && conf.Connectors != "{}\n" {
			values.Server.Config.Dex = conf.Connectors

			values.Dex = helm.DefaultDexValues(imageTags)

			if spec.NodeSelector != nil {
				for k, v := range spec.NodeSelector {
					values.Dex.NodeSelector[k] = v
				}
			}
			values.Server.Config.URL = strings.Join([]string{"https://", spec.Network.Domain}, "")
		}
	}

	if spec.AdditionalParameters != nil {
		if spec.AdditionalParameters.Server != nil {
			for _, param := range spec.AdditionalParameters.Server {
				values.Server.ExtraArgs = append(values.Server.ExtraArgs, param)
			}
		}
		if spec.AdditionalParameters.ApplicationController != nil {
			for _, param := range spec.AdditionalParameters.ApplicationController {
				values.Controller.ExtraArgs = append(values.Controller.ExtraArgs, param)
			}
		}
		if spec.AdditionalParameters.RepoServer != nil {
			for _, param := range spec.AdditionalParameters.RepoServer {
				values.RepoServer.ExtraArgs = append(values.RepoServer.ExtraArgs, param)
			}
		}
	}

	if spec.Rbac != nil {
		scopes := ""
		for _, scope := range spec.Rbac.Scopes {
			if scopes == "" {
				scopes = scope
			} else {
				scopes = strings.Join([]string{scopes, scope}, ", ")
			}
		}
		if scopes != "" {
			scopes = strings.Join([]string{"[", scopes, "]"}, "")
		}

		values.Server.RbacConfig = &helm.RbacConfig{
			Csv:     spec.Rbac.Csv,
			Default: spec.Rbac.Default,
			Scopes:  scopes,
		}
	}

	if spec.KnownHosts != nil && len(spec.KnownHosts) > 0 {
		knownHostsStr := values.Configs.KnownHosts.Data["ssh_known_hosts"]

		for _, v := range spec.KnownHosts {
			knownHostsStr = strings.Join([]string{knownHostsStr, v}, "\n")
		}

		values.Configs.KnownHosts.Data["ssh_known_hosts"] = knownHostsStr
	}

	if spec.NodeSelector != nil {
		for k, v := range spec.NodeSelector {
			values.Dex.NodeSelector[k] = v
			values.RepoServer.NodeSelector[k] = v
			values.Redis.NodeSelector[k] = v
			values.Controller.NodeSelector[k] = v
			values.Server.NodeSelector[k] = v
		}
	}

	if spec.Tolerations != nil {
		for _, tol := range spec.Tolerations {
			t := tol
			values.Dex.Tolerations = append(values.Dex.Tolerations, t)
			values.RepoServer.Tolerations = append(values.RepoServer.Tolerations, t)
			values.Redis.Tolerations = append(values.Redis.Tolerations, t)
			values.Controller.Tolerations = append(values.Controller.Tolerations, t)
			values.Server.Tolerations = append(values.Server.Tolerations, t)
		}
	}

	if spec.Redis != nil && spec.Redis.Resources != nil {
		values.Redis.Resources = spec.Redis.Resources
	}

	if spec.Dex != nil && spec.Dex.Resources != nil {
		values.Dex.Resources = spec.Dex.Resources
	}

	if spec.RepoServer != nil && spec.RepoServer.Resources != nil {
		values.RepoServer.Resources = spec.RepoServer.Resources
	}

	if spec.Server != nil && spec.Server.Resources != nil {
		values.Server.Resources = spec.Server.Resources
	}

	if spec.Controller != nil && spec.Controller.Resources != nil {
		values.Controller.Resources = spec.Controller.Resources
	}

	return values
}

func (a *Argocd) GetChartInfo() *chart.Chart {
	return helm.GetChartInfo()
}

func (a *Argocd) GetImageTags() map[string]string {
	return helm.GetImageTags()
}
