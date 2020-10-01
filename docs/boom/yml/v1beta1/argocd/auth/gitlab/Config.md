# Config 
 

## Structure 
 

| Attribute                  | Description                                                                                             | Default | Collection  |
| -------------------------- | ------------------------------------------------------------------------------------------------------- | ------- | ----------  |
| clientID                   | [here](secret/Secret.md)                                                                                |         |             |
| existingClientIDSecret     | Existing secret with the clientID , [here](secret/Existing.md)                                          |         |             |
| clientSecret               | [here](secret/Secret.md)                                                                                |         |             |
| existingClientSecretSecret | Existing secret with the clientSecret , [here](secret/Existing.md)                                      |         |             |
| baseURL                    | BaseURL of the gitlab instance                                                                          |         |             |
| groups                     | Optional groups whitelist, communicated through the "groups" scope                                      |         | X           |
| useLoginAsID               | Flag which will switch from using the internal GitLab id to the users handle (@mention) as the user id  |         |             |