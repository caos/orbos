# GopassStore 
 

## Structure 
 

| Attribute            | Description                                                                                            | Default | Collection | Map  |
| -------------------- | ------------------------------------------------------------------------------------------------------ | ------- | ---------- | ---  |
| sshKey               | [here](secret/Secret/Secret.md)                                                                        |         |            |      |
| existingSshKeySecret | Existing secret with ssh-key to clone the repository for gopass , [here](secret/Existing/Existing.md)  |         |            |      |
| gpgKey               | [here](secret/Secret/Secret.md)                                                                        |         |            |      |
| existingGpgKeySecret | Existing secret with gpg-key to decode the repository for gopass , [here](secret/Existing/Existing.md) |         |            |      |
| directory            | URL to repository for gopass store                                                                     |         |            |      |
| storeName            | Name of the gopass store                                                                               |         |            |      |