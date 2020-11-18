# Auth 
 

## Structure 
 

| Attribute                  | Description                                                                 | Default | Collection | Map  |
| -------------------------- | --------------------------------------------------------------------------- | ------- | ---------- | ---  |
| clientID                   | [here](secret/Secret/Secret.md)                                             |         |            |      |
| existingClientIDSecret     | Existing secret with the clientID , [here](secret/Existing/Existing.md)     |         |            |      |
| clientSecret               | [here](secret/Secret/Secret.md)                                             |         |            |      |
| existingClientSecretSecret | Existing secret with the clientSecret , [here](secret/Existing/Existing.md) |         |            |      |
| allowedOrganizations       | Organizations allowed to login                                              |         | X          |      |
| teamIDs                    | TeamIDs where the user is required to have at least one membership          |         | X          |      |