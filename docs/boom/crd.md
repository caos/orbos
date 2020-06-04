# CRD boom.caos.ch/v1beta1

## Structure

BOOM reconciles itself if a boomVersion is defined, if no boomVersion is defined there is no reconciling.

| Parameter                          | Description                                                                     | Default                           |
| ---------------------------------- | ------------------------------------------------------------------------------- | --------------------------------- |
| `boomVersion`                      | Version of BOOM which should be reconciled                                      |                                   |
| `currentStatePath`                 | Relative folder path where the currentstate is written to                       |                                   |
| `forceApply`                       | Flag if --force should be used by apply of resources                            |                                   |
| `preApply`                         | Spec for the yaml-files applied before applications                             |                                   |
| `postApply`                        | Spec for the yaml-files applied after applications                              |                                   |
| `prometheus-operator`              | Spec for the Prometheus-Operator                                                |                                   |
| `logging-operator`                 | Spec for the Banzaicloud Logging-Operator                                       |                                   |
| `prometheus-node-exporter`         | Spec for the Prometheus-Node-Exporter                                           |                                   |
| `prometheus-systemd-exporter`      | Spec for the Prometheus-Systemd-Exporter                                        |                                   |
| `grafana`                          | Spec for the Grafana                                                            |                                   |
| `ambassador`                       | Spec for the Ambassador                                                         |                                   |
| `kube-state-metrics`               | Spec for the Kube-State-Metrics                                                 |                                   |
| `argocd`                           | Spec for the Argo-CD                                                            |                                   |
| `prometheus`                       | Spec for the Prometheus instance                                                |                                   |
| `loki`                             | Spec for the Loki instance                                                      |                                   |

### Pre-Apply

| Parameter                          | Description                                                                     | Default                           |
| ---------------------------------- | ------------------------------------------------------------------------------- | --------------------------------- |
| `deploy`                           | Flag if tool should be deployed                                                 | false                             |
| `folder`                           | Relative path of folder in cloned git repository which should be applied        |                                   |

### Post-Apply

| Parameter                          | Description                                                                     | Default                           |
| ---------------------------------- | ------------------------------------------------------------------------------- | --------------------------------- |
| `deploy`                           | Flag if tool should be deployed                                                 | false                             |
| `folder`                           | Relative path of folder in cloned git repository which should be applied        |                                   |

### Prometheus-Operator

| Parameter                          | Description                                                                     | Default                           |
| ---------------------------------- | ------------------------------------------------------------------------------- | --------------------------------- |
| `deploy`                           | Flag if tool should be deployed                                                 | false                             |

### Logging-Operator

| Parameter                          | Description                                                                     | Default                           |
| ---------------------------------- | ------------------------------------------------------------------------------- | --------------------------------- |
| `deploy`                           | Flag if tool should be deployed                                                 | false                             |
| `fluentdStorage`                   | Spec to define how the persistency should be handled                            | nil                               |
| `fluentdStorage.size`              | Defined size of the PVC                                                         |                                   |
| `fluentdStorage.storageClass`      | Storageclass used by the PVC                                                    |                                   |
| `fluentdStorage.accessModes`       | Accessmodes used by the PVC                                                     |                                   |

### Prometheus-Node-Exporter

| Parameter                          | Description                                                                     | Default                           |
| ---------------------------------- | ------------------------------------------------------------------------------- | --------------------------------- |
| `deploy`                           | Flag if tool should be deployed                                                 | false                             |

### Grafana

| Parameter                          | Description                                                                     | Default                           |
| ---------------------------------- | ------------------------------------------------------------------------------- | --------------------------------- |
| `deploy`                           | Flag if tool should be deployed                                                 | false                             |
| `admin`                            | Spec for the definition of the admin account                                    | nil                               |
| `admin.existingSecret.name`        | Name of the secret which contains the admin account                             |                                   |
| `admin.existingSecret.idKey`       | Key of the username in the secret                                               |                                   |
| `admin.existingSecret.secretKey`   | Key of the password in the secret                                               |                                   |
| `admin`                            | Spec for the definition of the admin account                                    |                                   |
| `datasources`                      | Spec for additional datasources                                                 | nil                               |
| `datasources.name`                 | Name of the datasource                                                          |                                   |
| `datasources.type`                 | Type of the datasource (for example prometheus)                                 |                                   |
| `datasources.url`                  | URL to the datasource                                                           |                                   |
| `datasources.access`               | Access defintion of the datasource                                              |                                   |
| `datasources.isDefault`            | Boolean if datasource should be used as default                                 |                                   |
| `dashboardproviders`               | Spec for additional Dashboardproviders                                          | nil                               |
| `dashboardproviders.configMaps`    | ConfigMaps in which the dashboards are stored                                   |                                   |
| `dashboardproviders.folder`        | Local folder in which the dashboards are mounted                                |                                   |
| `storage`                          | Spec to define how the persistency should be handled                            | nil                               |
| `storage.size`                     | Defined size of the PVC                                                         |                                   |
| `storage.storageClass`             | Storageclass used by the PVC                                                    |                                   |
| `storage.accessModes`              | Accessmodes used by the PVC                                                     |                                   |
| `network`                          | Network configuration, [here](network.md)                                       |                                   |
| `auth`                             | Authorization and Authentication configuration for SSO, [here](sso-examples.md)  |                                   |

### Ambassador

| Parameter                          | Description                                                                     | Default                           |
| ---------------------------------- | ------------------------------------------------------------------------------- | --------------------------------- |
| `deploy`                           | Flag if tool should be deployed                                                 | false                             |
| `replicaCount`                     | Number of replicas used for deployment                                          | 1                                 |
| `service`                          | Service definition for ambassador                                               | nil                               |
| `service.type`                     | Type for the service                                                            | NodePort                          |
| `service.loadBalancerIP`           | Used IP for loadbalancing for ambassador if loadbalancer is used                | nil                               |

### Kube-State-Metrics

| Parameter                          | Description                                                                     | Default                           |
| ---------------------------------- | ------------------------------------------------------------------------------- | --------------------------------- |
| `deploy`                           | Flag if tool should be deployed                                                 | false                             |
| `replicaCount`                     | Number of replicas used for deployment                                          | 1                                 |

### Argo-CD

| Parameter                          | Description                                                                     | Default                           |
| ---------------------------------- | ------------------------------------------------------------------------------- | --------------------------------- |
| `deploy`                           | Flag if tool should be deployed                                                 | false                             |
| `customImage`                      | Custom argocd-image                                                             | nil                               |
| `customImage.enabled`              | Flag if custom argocd-image should get used with gopass                         | false                             |
| `customImage.imagePullSecret`      | Name of used imagePullSecret to pull customImage                                |                                   |
| `customImage.existingGpgKeySecret`         | Config to mount gpg key into repo-server pod                                    |                                   |
| `customImage.existingGpgKeySecret.name`    | Name of the existent secret which contains the gpg key                          |                                   |
| `customImage.existingGpgKeySecret.key`     | Key in the existent secret which contains the gpg key                           |                                   |
| `customImage.existingGpgKeySecret.internalName`    | Internal name used to mount the gpg key                                 |                                   |
| `customImage.existingSshKeySecret`         | Config to mount ssh key into repo-server pod                                    |                                   |
| `customImage.existingSshKeySecret.name`    | Name of the existent secret which contains the ssh key                          |                                   |
| `customImage.existingSshKeySecret.key`     | Key in the existent secret which contains the ssh key                           |                                   |
| `customImage.existingSshKeySecret.internalName`    | Internal name used to mount the ssh key                                 |                                   |
| `customImage.gopassDirectory`      | SSH-URL to Repository which is used as gopass secret store                      |                                   |
| `customImage.gopassStoreName`      | Name of the gopass secret store                                                 |                                   |
| `rbacConfig`                       | Config for RBAC in argocd                                                       | nil                               |
| `rbacConfig.policy.csv`            | Attribute policy.csv which goes into configmap argocd-rbac-cm                   |                                   |
| `rbacConfig.policy.default`        | Attribute policy.default which goes into configmap argocd-rbac-cm               |                                   |
| `rbacConfig.scopes`                | List of scopes which go into configmap argocd-rbac-cm                           |                                   |
| `network`                          | Network configuration, [here](network.md)                                       | nil                               |
| `auth`                             | Authorization and Authentication configuration for SSO, [here](sso-examples.md) | nil                               |
| `repositories`                     | Repositories used by argocd, [here](argocd-repositories.md)                     | nil                               |
| `credentials`                      | Credentials used by argocd, [here](argocd-credentials.md)                       | nil                               |
| `knownHosts`                       | List of known_hosts as strings for argocd                                       | nil                               |

### Prometheus

When the metrics spec is nil all metrics will get scraped.

| Parameter                          | Description                                                                     | Default                           |
| ---------------------------------- | ------------------------------------------------------------------------------- | --------------------------------- |
| `deploy`                           | Flag if tool should be deployed                                                 | false                             |
| `metrics`                          | Spec to define which metrics should get scraped                                 | nil                               |
| `metrics.ambassador`               | Bool if metrics should get scraped                                              | false                             |
| `metrics.argocd`                   | Bool if metrics should get scraped                                              | false                             |
| `metrics.kube-state-metrics`       | Bool if metrics should get scraped                                              | false                             |
| `metrics.prometheus-node-exporter` | Bool if metrics should get scraped                                              | false                             |
| `metrics.prometheus-systemd-exporter` | Bool if metrics should get scraped                                           | false                             |
| `metrics.api-server`               | Bool if metrics should get scraped                                              | false                             |
| `metrics.prometheus-operator`      | Bool if metrics should get scraped                                              | false                             |
| `metrics.logging-operator`         | Bool if metrics should get scraped                                              | false                             |
| `metrics.loki`                     | Bool if metrics should get scraped                                              | false                             |
| `storage`                          | Spec to define how the persistency should be handled                            | nil                               |
| `storage.size`                     | Defined size of the PVC                                                         |                                   |
| `storage.storageClass`             | Storageclass used by the PVC                                                    |                                   |
| `storage.accessModes`              | Accessmodes used by the PVC                                                     |                                   |

### Prometheus-Systemd-Exporter 

| Parameter                          | Description                                                                     | Default                           |
| ---------------------------------- | ------------------------------------------------------------------------------- | --------------------------------- |
| `deploy`                           | Flag if tool should be deployed                                                 | false                             |

### Loki

When the logs spec is nil all logs will get persisted in loki.

| Parameter                          | Description                                                                     | Default                           |
| ---------------------------------- | ------------------------------------------------------------------------------- | --------------------------------- |
| `deploy`                           | Flag if tool should be deployed                                                 | false                             |
| `logs`                             | Spec to define which logs will get persisted                                    | nil                               |
| `logs.ambassador`                  | Bool if logs will get persisted                                                 | false                             |
| `logs.argocd`                      | Bool if logs will get persisted                                                 | false                             |
| `logs.kube-state-metrics`          | Bool if logs will get persisted                                                 | false                             |
| `logs.prometheus-node-exporter`    | Bool if logs will get persisted                                                 | false                             |
| `logs.grafana`                     | Bool if logs will get persisted                                                 | false                             |
| `logs.prometheus-operator`         | Bool if logs will get persisted                                                 | false                             |
| `logs.logging-operator`            | Bool if logs will get persisted                                                 | false                             |
| `logs.loki`                        | Bool if logs will get persisted                                                 | false                             |
| `storage`                          | Spec to define how the persistency should be handled                            | nil                               |
| `storage.size`                     | Defined size of the PVC                                                         |                                   |
| `storage.storageClass`             | Storageclass used by the PVC                                                    |                                   |
| `storage.accessModes`              | Accessmodes used by the PVC                                                     |                                   |
