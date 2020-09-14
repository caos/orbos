# Repository 
 

 For a repository there are two types, with ssh-connection where an url and a certificate have to be provided and an https-connection where an URL, username and password have to be provided.


## Structure 
 

| Attribute                 | Description                                                                  | Default | Collection  |
| ------------------------- | ---------------------------------------------------------------------------- | ------- | ----------  |
| name                      | Internal used name                                                           |         |             |
| url                       | Prefix where the credential should be used (starting "git@" or "https://" )  |         |             |
| username                  | [here](secret/Secret.md)                                                     |         |             |
| existingUsernameSecret    | Existing secret used for username , [here](secret/Existing.md)               |         |             |
| password                  | [here](secret/Secret.md)                                                     |         |             |
| existingPasswordSecret    | Existing secret used for password , [here](secret/Existing.md)               |         |             |
| certificate               | [here](secret/Secret.md)                                                     |         |             |
| existingCertificateSecret | Existing secret used for certificate , [here](secret/Existing.md)            |         |             |