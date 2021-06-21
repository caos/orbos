# Pool 
 

## Structure 
 

| Attribute       | Description                                                                    | Default | Collection | Map  |
| --------------- | ------------------------------------------------------------------------------ | ------- | ---------- | ---  |
| osimage         | Used OS-image for the VMs in the pool                                          |         |            |      |
| mincpucores     | Minimum of requested v-CPU-cores for the VMs in the pool                       |         |            |      |
| minmemorygb     | Minimum of requested memory for the VMs in the pool                            |         |            |      |
| storagegb       | GB of storage requestes for the VMs in the pool                                |         |            |      |
| storageDiskType | Type of the used storage disk                                                  |         |            |      |
| preemptible     | Flag if VMs should be preemptible and can be shutdown and restarted after 24h  |         |            |      |
| localssds       | Count of mounted local SSDs with a size of 370 GB                              |         |            |      |