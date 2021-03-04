# Monitoring 
 

## Structure 
 

| Attribute          | Description                                                                                  | Default | Collection | Map  |
| ------------------ | -------------------------------------------------------------------------------------------- | ------- | ---------- | ---  |
| deploy             | Flag if tool should be deployed                                                              |  false  |            |      |
| admin              | Spec for the definition of the admin account , [here](admin/Admin/Admin.md)                  |         |            |      |
| datasources        | Spec for additional datasources , [here](Datasource/Datasource.md)                           |         | X          |      |
| dashboardproviders | Spec for additional Dashboardproviders , [here](Provider/Provider.md)                        |         | X          |      |
| storage            | Spec to define how the persistence should be handled , [here](storage/Spec/Spec.md)          |         |            |      |
| network            | Network configuration, used for SSO and external access , [here](network/Network/Network.md) |         |            |      |
| auth               | Authorization and Authentication configuration for SSO , [here](auth/Auth/Auth.md)           |         |            |      |
| plugins            | List of plugins which get added to the grafana instance                                      |         | X          |      |
| nodeSelector       | NodeSelector for deployment                                                                  |         |            | X    |
| tolerations        | Tolerations to run grafana on nodes , [here](k8s/Tolerations/Tolerations.md)                 |         | X          |      |
| resources          | Resource requirements , [here](k8s/Resources/Resources.md)                                   |         |            |      |
| overwriteVersion   | Override used image version                                                                  |         |            |      |