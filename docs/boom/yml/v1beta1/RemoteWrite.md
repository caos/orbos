# RemoteWrite 
 

## Structure 
 

| Attribute      | Description                                                                          | Default | Collection  |
| -------------- | ------------------------------------------------------------------------------------ | ------- | ----------  |
| url            | URL of the endpoint of the remote prometheus                                         |         |             |
| basicAuth      | Basic-auth-configuration to push metrics to remote prometheus , [here](BasicAuth.md) |         |             |
| relabelConfigs | RelabelConfigs for remote write , [here](RelabelConfig.md)                           |         | X           |