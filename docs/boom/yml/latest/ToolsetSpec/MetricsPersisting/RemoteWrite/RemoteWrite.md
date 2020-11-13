# RemoteWrite 
 

## Structure 
 

| Attribute      | Description                                                                                    | Default | Collection | Map  |
| -------------- | ---------------------------------------------------------------------------------------------- | ------- | ---------- | ---  |
| url            | URL of the endpoint of the remote prometheus                                                   |         |            |      |
| basicAuth      | Basic-auth-configuration to push metrics to remote prometheus , [here](BasicAuth/BasicAuth.md) |         |            |      |
| relabelConfigs | RelabelConfigs for remote write , [here](RelabelConfig/RelabelConfig.md)                       |         | X          |      |