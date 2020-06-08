# Repositories

For a repository there are two types, with ssh-connection where an url and a certifacte have to be provided and an https-connection where an URL, username and password have to be provided.

| Parameter                          | Description                                                                     | Default                           |
| ---------------------------------- | ------------------------------------------------------------------------------- | --------------------------------- |
| `url`                              | Prefix where the credential should be used (starting "git@" or "https://" )     |                                   |
| `existingUsernameSecret`           | Attributes for username in a secret                                             |                                   |
| `existingUsernameSecret.name`      | Name of the secret                                                              |                                   |
| `existingUsernameSecret.key`       | Key in the secret which contains the username                                   |                                   |
| `existingPasswordSecret`           | Attributes for username in a secret                                             |                                   |
| `existingPasswordSecret.name`      | Name of the secret                                                              |                                   |
| `existingPasswordSecret.key`       | Key in the secret which contains the password                                   |                                   |
| `existingCertificateSecret`        | Attributes for username in a secret                                             |                                   |
| `existingCertificateSecret.name`   | Name of the secret                                                              |                                   |
| `existingCertificateSecret.key`    | Key in the secret which contains the certificate                                |                                   |
