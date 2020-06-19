# OIDC 
 

## Structure 
 

| Attribute                  | Description                                                                                                | Default | Collection  |
| -------------------------- | ---------------------------------------------------------------------------------------------------------- | ------- | ----------  |
| name                       | Internal name of the OIDC provider                                                                         |         |             |
| issuer                     | Issuer of the OIDC provider                                                                                |         |             |
| clientID                   | [here](secret/Secret.md)                                                                                   |         |             |
| existingClientIDSecret     | Existing secret with the clientID , [here](secret/Existing.md)                                             |         |             |
| clientSecret               | [here](secret/Secret.md)                                                                                   |         |             |
| existingClientSecretSecret | Existing secret with the clientSecret , [here](secret/Existing.md)                                         |         |             |
| requestedScopes            | Optional set of OIDC scopes to request. If omitted, defaults to: ["openid", "profile", "email", "groups"]  |         | X           |
| requestedIDTokenClaims     | Optional set of OIDC claims to request on the ID token.                                                    |         |             |