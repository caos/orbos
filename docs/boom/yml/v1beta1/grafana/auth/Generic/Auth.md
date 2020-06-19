# Auth 
 

## Structure 
 

| Attribute                  | Description                                                        | Default | Collection  |
| -------------------------- | ------------------------------------------------------------------ | ------- | ----------  |
| clientID                   | [here](secret/Secret.md)                                           |         |             |
| existingClientIDSecret     | Existing secret with the clientID , [here](secret/Existing.md)     |         |             |
| clientSecret               | [here](secret/Secret.md)                                           |         |             |
| existingClientSecretSecret | Existing secret with the clientSecret , [here](secret/Existing.md) |         |             |
| scopes                     | Used scopes for the OAuth-flow                                     |         | X           |
| authURL                    | Auth-endpoint                                                      |         |             |
| tokenURL                   | Token-endpoint                                                     |         |             |
| apiURL                     | Userinfo-endpoint                                                  |         |             |
| allowedDomains             | Domains allowed to login                                           |         | X           |