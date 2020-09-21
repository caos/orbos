# Spec 
 

## Structure 
 

| Attribute      | Description                                                                            | Default | Collection | Map  |
| -------------- | -------------------------------------------------------------------------------------- | ------- | ---------- | ---  |
| verbose        | Flag to set log-level to debug                                                         |         |            |      |
| region         | Region used for all elements on GCE which are region specific                          |         |            |      |
| zone           | Zone used for all elements on GCE which are zone specific                              |         |            |      |
| pools          | List of Pools with an identification key which will get ensured , [here](Pool/Pool.md) |         |            | X    |
| sshkey         | SSH-key for connection to the VMs on GCE                                               |         |            |      |
| rebootRequired | List of nodes which are required to reboot                                             |         | X          |      |