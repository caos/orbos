# Userguides

## Kubernetes

### Create Cluster

1. Create `orb`
2. Create `orbiter.yml`
3. Write `secrets.yml`
4. Takeoff

(Re)Create the cluster:

```bash
orbctl -f {orbfile} takeoff
```

### Reset Cluster

If you would like to reset a cluster, in this example `kubernetes`, just follow this steps.

First we need to `destroy` it and afterwards we start over.

> !!! Danger Zone !!! This will reset all nodes including the `master` and `etcd` ones.

Destroy the cluster:

```bash
orbctl -f {orbfile} destroy
```

(Re)Create the cluster:

```bash
orbctl -f {orbfile} takeoff
```
