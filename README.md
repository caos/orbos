This project is in alpha state. The API will continue breaking until version 1.0.0 is released
-----  

# Orbiter: The Meta Cluster Manager

[![semantic-release](https://img.shields.io/badge/%20%20%F0%9F%93%A6%F0%9F%9A%80-semantic--release-e10079.svg)](https://github.com/semantic-release/semantic-release) ![](https://github.com/caos/orbiter/workflows/Release/badge.svg)


Orbiter boostraps, lifecycles and destroys clustered software and other cluster managers whereas each can be configured to span over a wide range of infrastructure providers.

## Bootstrap a new static cluster on firecracker VMs using ignite

Create a new repository (e.g. git@github.com:caos/my-orb.git)  

```bash
# Download latest orbctl
curl -s https://api.github.com/repos/caos/orbiter/releases/latest | grep "browser_download_url.*orbctl-$(uname)-$(uname -m)" | cut -d '"' -f 4 | sudo wget -i - -O /usr/local/bin/orbctl -
sudo chmod +x /usr/local/bin/orbctl
sudo chown $(id -u):$(id -g) /usr/local/bin/orbctl

# Create a new ssh key pair.
ssh-keygen -t rsa -b 4096 -C "repo and VM bootstrap key" -P "" -f "~/.ssh/myorb_bootstrap" -q && ssh-add "~/.ssh/myorb_bootstrap"

# Create a new orbconfig
mkdir -p ~/.orb
cat > ~/.orb/config << EOF
url: git@github.com:me/my-orb.git
masterkey: a very secret key
repokey: |
$(cat ~/.ssh/myorb_bootstrap | sed s/^/\ \ /g)
EOF
```

Add the public part to the git repositories trusted deploy keys.  

```bash
# Add the bootstrap key pair to the remote secrets file. For simplicity, we use the repokey here.
orbctl addsecret myorbprodstatic_bootstrapkey --stdin
orbctl addsecret myorbprodstatic_bootstrapkey_pub --stdin

# Create four firecracker VMs
sudo ignite run weaveworks/ignite-ubuntu --cpus 2 --memory 4GB --size 15GB --ssh=~/.ssh/myorb_bootstrap.pub --ports 5000:5000 --ports 6443:6443 --name first
sudo ignite run weaveworks/ignite-ubuntu --cpus 2 --memory 4GB --size 15GB --ssh=~/.ssh/myorb_bootstrap.pub --ports 5000:5000 --ports 6443:6443 --name second
sudo ignite run weaveworks/ignite-ubuntu --cpus 2 --memory 4GB --size 15GB --ssh=~/.ssh/myorb_bootstrap.pub --ports 5000:5000 --ports 6443:6443 --name third
sudo ignite run weaveworks/ignite-ubuntu --cpus 2 --memory 4GB --size 15GB --ssh=~/.ssh/myorb_bootstrap.pub --ports 5000:5000 --ports 6443:6443 --name fourth
```

Make sure your orb repo contains a desired.yml file similar to [this example](examples/dayone/desired.yml). Show your VMs IPs with `sudo ignite ps -a`  

```bash
# Your environment is ready now, finally we can do some actual work. Launch a local orbiter that bootstraps your orb
myorb takeoff

# When the orbiter exits, overwrite your kubeconfig by the newly created admin kubeconfig
mkdir -p ~/.kube && myorb readsecret myorbprod_kubeconfig > ~/.kube/config

# TODO: Not needed anymore when docker registry is anonymously pullable #39
kubectl -n kube-system create secret docker-registry orbiterregistry --docker-server=docker.pkg.github.com --docker-username=
${GITHUB_USERNAME} --docker-password=${GITHUB_ACCESS_TOKEN}

# Watch your nodes become ready
kubectl get nodes --watch

# Watch your pods become ready
kubectl get pods --all-namespaces --watch

# [Optional] Teach your ssh daemon to use the newly created ssh key for connecting to the VMS directly. The bootstrap key is not going to work anymore. 
myorb readsecret myorbprodstatic_maintenancekey > ~/.ssh/myorb-maintenance && chmod 0600 ~/.ssh/myorb-maintenance && ssh-add ~/.ssh/myorb-maintenance

# Cleanup your environment
sudo ignite rm -f $(sudo ignite ps -aq)
```

Delete your git repository

## Why another cluster manager?

We observe a universal trend of increasing system distribution. Key drivers are cloud native engineering, microservices architectures, global competition among hyperscalers and so on.

We embrace this trend but counteract its biggest downside, the associated increase of complexity in managing all these distributed systems. Our goal is to enable players of any size to run clusters of any type using infrastructure from any provider. Orbiter is a tool to do this in a reliable, secure, auditable, cost efficient way, preventing vendor lock-in, monoliths consisting of microservices and human failure doing repetitive tasks.

What makes Orbiter special is that it ships with a nice **Launchpad UI** providing useful tools to interact intuitively with the operator. Also, the operational design follows the **GitOps pattern**, highlighting day two operations, sticking to a distinct source of truth for declarative system configuration and maintaining a consistent audit log, everything out-of-the-box. All managed software can be configured to be **self updating** according to special policies, including Orbiter itself. Then, the Orbiter code base is designed to be **highly extendable**, which ensures that any given tool can eventually run on any desired provider.

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

