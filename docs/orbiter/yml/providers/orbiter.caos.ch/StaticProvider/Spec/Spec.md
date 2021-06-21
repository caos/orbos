# Spec 
 

## Structure 
 

| Attribute          | Description                                                                                  | Default | Collection | Map  |
| ------------------ | -------------------------------------------------------------------------------------------- | ------- | ---------- | ---  |
| verbose            | Flag to set log-level to debug                                                               |         |            |      |
| pools              | List of Pools with an identification key which will get ensured , [here](Machine/Machine.md) |         | X          | X    |
| keys               | Used SSH-keys used for ensuring , [here](Keys/Keys.md)                                       |         |            |      |
| externalInterfaces |                                                                                              |         | X          |      |