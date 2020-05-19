package argocd

import (
	"github.com/caos/orbos/internal/operator/boom/application/applications/argocd/config/credential"
	"github.com/caos/orbos/internal/operator/boom/application/applications/argocd/config/repository"
	"strings"

	toolsetsv1beta1 "github.com/caos/orbos/internal/operator/boom/api/v1beta1"
	"github.com/caos/orbos/internal/operator/boom/application/applications/argocd/config"
	"github.com/caos/orbos/internal/operator/boom/application/applications/argocd/customimage"
	"github.com/caos/orbos/internal/operator/boom/application/applications/argocd/helm"
	"github.com/caos/orbos/internal/operator/boom/templator/helm/chart"
	"github.com/caos/orbos/mntr"
)

func (a *Argocd) HelmPreApplySteps(monitor mntr.Monitor, toolsetCRDSpec *toolsetsv1beta1.ToolsetSpec) ([]interface{}, error) {
	addedSecrets := customimage.GetSecrets(toolsetCRDSpec.Argocd)
	repoSecrets := repository.GetSecrets(toolsetCRDSpec.Argocd)
	credSecrets := credential.GetSecrets(toolsetCRDSpec.Argocd)

	addedSecrets = append(addedSecrets, repoSecrets...)
	addedSecrets = append(addedSecrets, credSecrets...)
	return addedSecrets, nil
}

func (a *Argocd) HelmMutate(monitor mntr.Monitor, toolsetCRDSpec *toolsetsv1beta1.ToolsetSpec, resultFilePath string) error {
	spec := toolsetCRDSpec.Argocd

	if spec.CustomImage != nil && spec.CustomImage.Enabled && spec.CustomImage.ImagePullSecret != "" {
		if err := customimage.AddImagePullSecretFromSpec(spec, resultFilePath); err != nil {
			return err
		}

		if spec.CustomImage.GopassStores != nil && len(spec.CustomImage.GopassStores) > 0 {
			if err := customimage.AddPostStartFromSpec(spec, resultFilePath); err != nil {
				return err
			}
		}
	}

	return nil
}

func (a *Argocd) SpecToHelmValues(monitor mntr.Monitor, toolsetCRDSpec *toolsetsv1beta1.ToolsetSpec) interface{} {
	spec := toolsetCRDSpec.Argocd

	imageTags := a.GetImageTags()
	values := helm.DefaultValues(imageTags)
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
			values.Server.Config.URL = strings.Join([]string{"https://", spec.Network.Domain}, "")
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

	return values
}

func (a *Argocd) GetChartInfo() *chart.Chart {
	return helm.GetChartInfo()
}

func (a *Argocd) GetImageTags() map[string]string {
	return helm.GetImageTags()
}
