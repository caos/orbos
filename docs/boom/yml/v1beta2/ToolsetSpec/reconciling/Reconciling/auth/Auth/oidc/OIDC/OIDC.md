# OIDC 
 

## Structure 
 

| Attribute                  | Description                                                                                                | Default | Collection | Map  |
| -------------------------- | ---------------------------------------------------------------------------------------------------------- | ------- | ---------- | ---  |
| name                       | Internal name of the OIDC provider                                                                         |         |            |      |
| issuer                     | Issuer of the OIDC provider                                                                                |         |            |      |
| clientID                   | [here](secret/Secret/Secret.md)                                                                            |         |            |      |
| existingClientIDSecret     | Existing secret with the clientID , [here](secret/Existing/Existing.md)                                    |         |            |      |
| clientSecret               | [here](secret/Secret/Secret.md)                                                                            |         |            |      |
| existingClientSecretSecret | Existing secret with the clientSecret , [here](secret/Existing/Existing.md)                                |         |            |      |
| requestedScopes            | Optional set of OIDC scopes to request. If omitted, defaults to: ["openid", "profile", "email", "groups"]  |         | X          |      |
| requestedIDTokenClaims     | Optional set of OIDC claims to request on the ID token. , [here](Claim/Claim.md)                           |         |            | X    |