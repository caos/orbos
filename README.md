This project is in alpha state. The API will continue breaking until version 1.0.0 is released
-----  

# Orbiter: The Meta Cluster Manager

[![semantic-release](https://img.shields.io/badge/%20%20%F0%9F%93%A6%F0%9F%9A%80-semantic--release-e10079.svg)](https://github.com/semantic-release/semantic-release) ![](https://github.com/caos/orbiter/workflows/Release/badge.svg)


Orbiter boostraps, lifecycles and destroys clustered software and other cluster managers whereas each can be configured to span over a wide range of infrastructure providers.

## Bootstrap a new static cluster on firecracker VMs using ignite

1. Create a new repository (e.g. git@github.com:caos/orbiter-tmp.git), clone it and change directory to its root and export according environment variables.
```bash
export ORBITER_REPOSITORY=git@github.com:caos/orbiter-tmp.git
# Set ORBITER_SECRETSPREFIX to your repositorys name without dashes
export ORBITER_SECRETSPREFIX=orbitertmp
```

1. Initialize Orbiter runtime secrets
```bash
# Create a master key used to symmetrically encrypt all other keys
sudo mkdir -p /etc/orbiter && sudo chown $(id -u):$(id -g) /etc/orbiter && echo -n "a very secret key!" > /etc/orbiter/masterkey && chmod 600 /etc/orbiter/masterkey

# Create a new key pair. This will be used to bootstrap new VMs as well as authenticating our git calls.
ssh-keygen -t rsa -b 4096 -C "repo and VM bootstrap key" -P "" -f "/etc/orbiter/repokey" -q

# Add the private part to your ssh daemon
ssh-add /etc/orbiter/repokey

# Add the bootstrap private key
cat /etc/orbiter/repokey | docker run --rm --user $(id -u):$(id -g) --volume $(pwd):/secrets --volume /etc/orbiter:/etc/orbiter:ro --workdir /secrets --interactive docker.pkg.github.com/caos/orbiter/orbiter:latest --addsecret ${ORBITER_SECRETSPREFIX}prodstatic_bootstrapkey

# Add the bootstrap public key
cat /etc/orbiter/repokey.pub | docker run --rm --user $(id -u):$(id -g) --volume $(pwd):/secrets --volume /etc/orbiter:/etc/orbiter:ro --workdir /secrets --interactive docker.pkg.github.com/caos/orbiter/orbiter:latest --addsecret ${ORBITER_SECRETSPREFIX}prodstatic_bootstrapkey_pub
```

1. Add your generated public key to your git repositorys deploy keys with write access

1. Create four firecracker VMs
```bash
sudo ignite run weaveworks/ignite-ubuntu --cpus 2 --memory 4GB --size 15GB --ssh=/etc/orbiter/repokey.pub --ports 5000:5000 --ports 6443:6443 --name first
sudo ignite run weaveworks/ignite-ubuntu --cpus 2 --memory 4GB --size 15GB --ssh=/etc/orbiter/repokey.pub --ports 5000:5000 --ports 6443:6443 --name second
sudo ignite run weaveworks/ignite-ubuntu --cpus 2 --memory 4GB --size 15GB --ssh=/etc/orbiter/repokey.pub --ports 5000:5000 --ports 6443:6443 --name third
sudo ignite run weaveworks/ignite-ubuntu --cpus 2 --memory 4GB --size 15GB --ssh=/etc/orbiter/repokey.pub --ports 5000:5000 --ports 6443:6443 --name fourth
```
1. Create a new file called desired.yml (see [](examples/dayone/desired.yml)). Replace the IPs in the example file by the outputs from `sudo ignite ps -a`
1. Push your newly created yaml files `git add . && git commit -m "bootstrap cluster" && git push`.
1. Bootstrap a cluster. Adjust the arguments to your setup
```bash
docker run --rm --volume /etc/orbiter:/etc/orbiter:ro --user $(id -u):$(id -g) docker.pkg.github.com/caos/orbiter/orbiter:latest --repourl $ORBITER_REPOSITORY
```
1. Connect to the cluster by using the automatically created new secrets
```bash
# Update git changes made by Orbiter
git pull

# Teach your ssh daemon to use the newly created ssh key for connecting the VMS directly. The bootstrap key is not going to work anymore. 
docker run --rm --user $(id -u):$(id -g) --volume $(pwd):/secrets --volume /etc/orbiter:/etc/orbiter:ro --workdir /secrets --interactive docker.pkg.github.com/caos/orbiter/orbiter:latest --readsecret ${ORBITER_SECRETSPREFIX}prodstatic_maintenancekey > /tmp/orbiter-maintenancekey && chmod 0600 /tmp/orbiter-maintenancekey && ssh-add /tmp/orbiter-maintenancekey

# Overwrite your kubeconfig by the newly created admin kubeconfig
mkdir -p ~/.kube && docker run --rm --user $(id -u):$(id -g) --volume $(pwd):/secrets --volume /etc/orbiter:/etc/orbiter:ro --workdir /secrets --interactive docker.pkg.github.com/caos/orbiter/orbiter:latest --readsecret ${ORBITER_SECRETSPREFIX}prod_kubeconfig > ~/.kube/config

# TODO: Not needed anymore when docker registry is public for reading #39
kubectl -n kube-system create secret docker-registry orbiterregistry --docker-server=docker.pkg.github.com --docker-username=${GITHUB_USERNAME} --docker-password=${GITHUB_ACCESS_TOKEN}

# Watch your nodes becoming ready
kubectl get nodes --watch

# Watch your pods becoming ready
kubectl get pods --all-namespaces --watch
```

1. Overwrite your desired.yml by the contents of examples/daytwo/desired.yml, push your changes with `git add . && git commit -m "change cluster" && git push` and let Orbiter do its work.
1. Cleanup your environment
```bash
sudo ignite rm -f $(sudo ignite ps -aq)
```

## Why another cluster manager?

We observe a universal trend of increasing system distribution. Key drivers are cloud native engineering, microservices architectures, global competition among hyperscalers and so on.

We embrace this trend but counteract its biggest downside, the associated increase of complexity in managing all these distributed systems. Our goal is to enable players of any size to run clusters of any type using infrastructure from any provider. Orbiter is a tool to do this in a reliable, secure, auditable, cost efficient way, preventing vendor lock-in, monoliths consisting of microservices and human failure doing repetitive tasks.

What makes Orbiter special is that it ships with a nice **Launchpad UI** providing useful tools to interact intuitively with the operator. Also, the operational design follows the **GitOps pattern**, highlighting day two operations, sticking to a distinct source of truth for declarative system configuration and maintaining a consistent audit log, everything out-of-the-box. All managed Software can be configured to be **self updating** according to special policies, including Orbiter itself. Then, the Orbiter code base is designed to be **highly extendable**, which ensures that any given tool can eventually run on any desired provider.

## Supported clusters

- Kubernetes (vanilla)

More to come:
- Other cluster managers
    - Nomad
    - Mesos
    - YARN
    - ...
- Databases
    - etcd
    - Zookeeper
    - CockroachDB
    - Aerospike
    - Redis
    - ...
- Message Brokers
    - Kafka
    - RabbitMQ
    - VerneMQ
    - ...
- ...

If you desire an explicit implementation, file an issue. Also, pull requests are welcome and appreciated.

## Supported providers

- Google Compute Engine
- Static provider (orbiter only manages clusters, infrastructure is already existing and managed manually)

More to come:
- Hyperscalers
    - Amazon Web Services
    - Alibaba Cloud
    - Microsoft Azure
    - Oracle Cloud
    - IBM Cloud
    - ...
- Virtualization software
    - VMWare
    - KubeVirt
    - ...
- Bare Metal
    - PXE-Boot
    - ...
- ... 

If you desire an explicit implementation, file an issue. Also, pull requests are welcome and appreciated.

## How does it work?

An Orbiter instance runs in a Kubernetes Pod managing n configured clusters, typically including the one it is running in. It scales the clusters nodes and instructs Node Agents over the kube-apiserver which software to install on the node they run on. The Node Agents run as native system processes which are managed by systemd.

For more details, take a look at the [technical docs](./docs/README.md).

## How to develop?

Configure your tooling to use certain environment variables. E.g. in VSCode, add the following to your settings.json.
```json
{
    "go.testEnvVars": {
        "MODE": "DEBUG",
        "ORBITER_ROOT": "/home/elio/Code/src/github.com/caos/orbiter"
    },
    "go.testTimeout": "40m",
}
```

Run the tests you find in internal/kinds/clusters/kubernetes/test/kubernetes_test.go in debug mode

For debugging node agents, use a configuration similar to the following VSCode launch.json, adjusting the host IP
```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "nodeagent",
            "type": "go",
            "request": "attach",
            "apiVersion": 2,
            "mode": "remote",
            "port": 5000,
            "host": "10.61.0.127"
        },
    ]
}
```
## License

The full functionality of the operator is and stays open source and free to use for everyone. We pay our wages by using Orbiter for selling further workload enterprise services like support, monitoring and forecasting, IAM, CI/CD, secrets management etc. Visit our [website](https://caos.ch) and get in touch.

See the exact licensing terms [here](./LICENSE)

