# Argocd 
 

## Structure 
 

| Attribute    | Description                                                                                  | Default | Collection | Map  |
| ------------ | -------------------------------------------------------------------------------------------- | ------- | ---------- | ---  |
| deploy       | Flag if tool should be deployed                                                              |  false  |            |      |
| customImage  | Use of custom argocd-image which includes gopass , [here](CustomImage/CustomImage.md)        |  false  |            |      |
| network      | Network configuration, used for SSO and external access , [here](network/Network/Network.md) |         |            |      |
| auth         | Authorization and Authentication configuration for SSO , [here](auth/Auth/Auth.md)           |         |            |      |
| rbacConfig   | Configuration for RBAC in argocd , [here](Rbac/Rbac.md)                                      |         |            |      |
| repositories | Repositories used by argocd , [here](repository/Repository/Repository.md)                    |         | X          |      |
| credentials  | Credentials used by argocd , [here](repository/Repository/Repository.md)                     |         | X          |      |
| knownHosts   | List of known_hosts as strings for argocd                                                    |         | X          |      |
| nodeSelector | NodeSelector for deployment                                                                  |         |            | X    |