# ExternalConfig 
 

## Structure 
 

| Attribute            | Description                                                                           | Default | Collection | Map  |
| -------------------- | ------------------------------------------------------------------------------------- | ------- | ---------- | ---  |
| verbose              | Verbose flag to set debug-level to debug                                              |         |            |      |
| domain               | Domain used on cloudflare                                                             |         |            |      |
| iP                   | IP used for all DNS entries                                                           |         |            |      |
| rules                | List of firewall rules , [here](Rule/Rule.md)                                         |         | X          |      |
| groups               | List of group definition which can be used in firewall rules , [here](Group/Group.md) |         | X          |      |
| credentials          | Credentials used for all actions with cloudflare , [here](Credentials/Credentials.md) |         |            |      |
| prefix               | Prefix given to firewall rules descriptions                                           |         |            |      |
| additionalSubdomains | Additional DNS entries besides the ones for zitadel , [here](Subdomain/Subdomain.md)  |         | X          |      |