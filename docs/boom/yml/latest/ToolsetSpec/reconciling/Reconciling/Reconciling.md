# Reconciling 
 

## Structure 
 

| Attribute            | Description                                                                                            | Default | Collection | Map  |
| -------------------- | ------------------------------------------------------------------------------------------------------ | ------- | ---------- | ---  |
| deploy               | Flag if tool should be deployed                                                                        |  false  |            |      |
| customImage          | Use of custom argocd-image which includes gopass , [here](CustomImage/CustomImage.md)                  |         |            |      |
| network              | Network configuration, used for SSO and external access , [here](network/Network/Network.md)           |         |            |      |
| auth                 | Authorization and Authentication configuration for SSO , [here](auth/Auth/Auth.md)                     |         |            |      |
| rbacConfig           | Configuration for RBAC in argocd , [here](Rbac/Rbac.md)                                                |         |            |      |
| repositories         | Repositories used by argocd , [here](repository/Repository/Repository.md)                              |         | X          |      |
| credentials          | Credentials used by argocd , [here](repository/Repository/Repository.md)                               |         | X          |      |
| knownHosts           | List of known_hosts as strings for argocd                                                              |         | X          |      |
| nodeSelector         | NodeSelector for deployment                                                                            |         |            | X    |
| tolerations          | Tolerations to run argocd on nodes , [here](k8s/Tolerations/Tolerations.md)                            |         | X          |      |
| dex                  | Dex options , [here](CommonComponent/CommonComponent.md)                                               |         |            |      |
| repoServer           | RepoServer options , [here](CommonComponent/CommonComponent.md)                                        |         |            |      |
| redis                | Redis options , [here](CommonComponent/CommonComponent.md)                                             |         |            |      |
| controller           | Controller options , [here](CommonComponent/CommonComponent.md)                                        |         |            |      |
| server               | Server options , [here](CommonComponent/CommonComponent.md)                                            |         |            |      |
| overwriteVersion     | Override used image version                                                                            |         |            |      |
| additionalParameters | Additional parameters to use in the deployments , [here](AdditionalParameters/AdditionalParameters.md) |         |            |      |