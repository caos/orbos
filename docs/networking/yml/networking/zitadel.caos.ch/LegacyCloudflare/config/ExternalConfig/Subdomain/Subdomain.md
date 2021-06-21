# Subdomain 
 

## Structure 
 

| Attribute | Description                                                                                                                                                                             | Default | Collection | Map  |
| --------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------- | ---------- | ---  |
| subdomain | Name of the subdomain                                                                                                                                                                   |         |            |      |
| ip        | IP which is the target of the DNS entry                                                                                                                                                 |         |            |      |
| proxied   | Flag if DNS entry is proxied by cloudflare                                                                                                                                              |         |            |      |
| ttl       | Time-to-live for the DNS entry                                                                                                                                                          |         |            |      |
| type      | Type of the DNS entry                                                                                                                                                                   |         |            |      |
| priority  | The priority of the rule to allow control of processing order. A lower number indicates high priority. If not provided, any rules with a priority will be sequenced before those witho  |         |            |      |