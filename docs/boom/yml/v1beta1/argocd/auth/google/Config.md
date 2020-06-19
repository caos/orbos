# Config 
 

## Structure 
 

| Attribute                        | Description                                                                        | Default | Collection  |
| -------------------------------- | ---------------------------------------------------------------------------------- | ------- | ----------  |
| clientID                         | [here](secret/Secret.md)                                                           |         |             |
| existingClientIDSecret           | Existing secret with the clientID , [here](secret/Existing.md)                     |         |             |
| clientSecret                     | [here](secret/Secret.md)                                                           |         |             |
| existingClientSecretSecret       | Existing secret with the clientSecret , [here](secret/Existing.md)                 |         |             |
| hostedDomains                    | List of hosted domains which are permitted to login                                |         | X           |
| groups                           | List of groups in the hosted domains which are permitted to login                  |         | X           |
| serviceAccountJSON               | [here](secret/Secret.md)                                                           |         |             |
| existingServiceAccountJSONSecret | Existing secret with the JSON of the service account , [here](secret/Existing.md)  |         |             |
| serviceAccountFilePath           | File where the serviceAccountJSON will get persisted to impersonate G Suite admin  |         |             |
| adminEmail                       | Email of a G Suite admin to impersonate                                            |         |             |