# Userguides

## Kubernetes

### Create Cluster

#### Create orb.yml

[orb.yml](../examples/k8s/orb.yml)

#### Create orbiter.yml

[orbiter.yml](../examples/k8s/static/orbiter.yml)

#### Write Secrets

`orbctl -f orb.yml writesecret`

#### Takeoff

```bash
orbctl -f {orbfile} takeoff
```

### Tear down Cluster

If you would like to reset a cluster, in this example `kubernetes`, follow these steps.

First we need to `destroy` it and afterwards we start over.

> !!! Danger Zone !!! This will reset all nodes including the `master` and `etcd` ones.

Destroy the cluster:

```bash
orbctl -f {orbfile} destroy
```
