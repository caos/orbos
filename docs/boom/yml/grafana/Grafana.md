# Grafana 
 

## Structure 
 

| Attribute          | Description                                                                          | Default | Collection  |
| ------------------ | ------------------------------------------------------------------------------------ | ------- | ----------  |
| deploy             | Flag if tool should be deployed                                                      |  false  |             |
| admin              | Spec for the definition of the admin account , [here](admin/Admin.md)                |         |             |
| datasources        | Spec for additional datasources , [here](Datasource.md)                              |         | X           |
| dashboardproviders | Spec for additional Dashboardproviders , [here](Provider.md)                         |         | X           |
| storage            | Spec to define how the persistence should be handled , [here](storage/Spec.md)       |         |             |
| network            | Network configuration, used for SSO and external access , [here](network/Network.md) |         |             |
| auth               | Authorization and Authentication configuration for SSO , [here](auth/Auth.md)        |         |             |