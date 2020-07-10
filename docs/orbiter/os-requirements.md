# Operating System Requirements

## System

- CentOS 7
- SSH daemon running
- Ability for Node Agent to disable swap (e.g. containers on a host with swap enabled won't work)
- For Kubernetes Clusters, a minimum of 2 CPU cores is required per node

## User
- orbiter user with passwordless sudo capability
- Bootstrapkey listed in /home/orbiter/.ssh/authorized_keys
