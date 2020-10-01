# GopassStore 
 

## Structure 
 

| Attribute            | Description                                                                                   | Default | Collection  |
| -------------------- | --------------------------------------------------------------------------------------------- | ------- | ----------  |
| sshKey               | [here](secret/Secret.md)                                                                      |         |             |
| existingSshKeySecret | Existing secret with ssh-key to clone the repository for gopass , [here](secret/Existing.md)  |         |             |
| gpgKey               | [here](secret/Secret.md)                                                                      |         |             |
| existingGpgKeySecret | Existing secret with gpg-key to decode the repository for gopass , [here](secret/Existing.md) |         |             |
| directory            | URL to repository for gopass store                                                            |         |             |
| storeName            | Name of the gopass store                                                                      |         |             |