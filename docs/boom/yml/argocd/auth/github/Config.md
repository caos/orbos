# Config 
 

## Structure 
 

| Attribute                  | Description                                                                                             | Default | Collection  |
| -------------------------- | ------------------------------------------------------------------------------------------------------- | ------- | ----------  |
| clientID                   | [here](secret/Secret.md)                                                                                |         |             |
| existingClientIDSecret     | Existing secret with the clientID , [here](secret/Existing.md)                                          |         |             |
| clientSecret               | [here](secret/Secret.md)                                                                                |         |             |
| existingClientSecretSecret | Existing secret with the clientSecret , [here](secret/Existing.md)                                      |         |             |
| orgs                       | Required membership to organization in github , [here](Org.md)                                          |         | X           |
| loadAllGroups              | Flag which indicates that all user groups and teams should be loaded                                    |         |             |
| teamNameField              | Optional choice between 'name' (default), 'slug', or 'both'                                             |         |             |
| useLoginAsID               | Flag which will switch from using the internal GitHub id to the users handle (@mention) as the user id  |         |             |