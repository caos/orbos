# Orbiter The Meta Cluster Manager

![semantic-release](https://img.shields.io/badge/%20%20%F0%9F%93%A6%F0%9F%9A%80-semantic--release-e10079.svg)
![Github Release Badge](https://github.com/caos/orbiter/workflows/Release/badge.svg)
[![GitHub license](https://img.shields.io/github/license/caos/orbiter)](https://github.com/caos/orbiter/blob/master/LICENSE)
[![GitHub release](https://img.shields.io/github/release/caos/orbiter)](https://GitHub.com/caos/orbiter/releases/)

> This project is in alpha state. The API will continue breaking until version 1.0.0 is released

## What Is It

`Orbiter` boostraps, lifecycles and destroys clustered software and other cluster managers whereas each can be configured to span over a wide range of infrastructure providers. Its focus is laid on automating away all `day two` operations, as we consider them to have much bigger impacts than `day one` operations from a business perspective.

## How Does It Work

An Orbiter instance runs as a Kubernetes Pod managing the configured clusters (i.e. an Orb), typically including the one it is running on. It scales the clusters nodes and has `Node Agents` install software packages on their operating systems. `Node Agents` run as native system processes managed by `systemd`. An Orbs Git repository is the only source of truth for desired state. Also, the current Orbs state is continously pushed to its Git repository, so not only changes to the desired state is always tracked but also the most important changes to the actual systems state.

For more details, take a look at the [design docs](./docs/design.md).

## Why Another Cluster Manager

We observe a universal trend of increasing system distribution. Key drivers are cloud native engineering, microservices architectures, global competition among hyperscalers and so on.

We embrace this trend but counteract its biggest downside, the associated increase of complexity in managing all these distributed systems. Our goal is to enable players of any size to run clusters of any type using infrastructure from any provider. Orbiter is a tool to do this in a reliable, secure, auditable, cost efficient way, preventing vendor lock-in, monoliths consisting of microservices and human failure doing repetitive tasks.

What makes Orbiter special is that it ships with a nice **Mission Control UI** (currently in closed alpha) providing useful tools to interact intuitively with the operator. Also, the operational design follows the **GitOps pattern**, highlighting `day two operations`, sticking to a distinct source of truth for declarative system configuration and maintaining a consistent audit log, everything out-of-the-box. Then, the Orbiter code base is designed to be **highly extendable**, which ensures that any given cluster type can eventually run on any desired provider.

## How To Use It

In the following example we will create a `kubernetes` cluster on a `static provider`. A `static provider` is a provider, which has no or little API for automation, e.g legacy VM's or Bare Metal scenarios.

### Download Orbctl

```bash
# Download latest orbctl
curl -s https://api.github.com/repos/caos/orbiter/releases/latest | grep "browser_download_url.*orbctl-$(uname)-$(uname -m)" | cut -d '"' -f 4 | sudo wget -i - -O /usr/local/bin/orbctl
sudo chmod +x /usr/local/bin/orbctl
sudo chown $(id -u):$(id -g) /usr/local/bin/orbctl
```

### Create Config And Secrets

```bash
# Create a new ssh key pair.
mkdir -p ~/.ssh && ssh-keygen -t rsa -b 4096 -C "repo and VM bootstrap key" -P "" -f ~/.ssh/myorb_bootstrap -q && ssh-add ~/.ssh/myorb_bootstrap

# Create a new orbconfig
mkdir -p ~/.orb
cat > ~/.orb/config << EOF
url: git@github.com:me/my-orb.git
masterkey: a very secret key
repokey: |
$(cat ~/.ssh/myorb_bootstrap | sed s/^/\ \ /g)
EOF
```

> Add the public part to the git repositories trusted deploy keys.  

```bash
# Add the bootstrap key pair to the remote secrets file. For simplicity, we use the repokey here.
orbctl addsecret myorbprodstatic_bootstrapkey --stdin
orbctl addsecret myorbprodstatic_bootstrapkey_pub --stdin
```

### Bootstrap A New Static Cluster On Firecracker VMs Using Ignite

```bash
# Create four firecracker VMs
sudo ignite run weaveworks/ignite-ubuntu --cpus 2 --memory 4GB --size 15GB --ssh=~/.ssh/myorb_bootstrap.pub --ports 5000:5000 --ports 6443:6443 --name first
sudo ignite run weaveworks/ignite-ubuntu --cpus 2 --memory 4GB --size 15GB --ssh=~/.ssh/myorb_bootstrap.pub --ports 5000:5000 --ports 6443:6443 --name second
sudo ignite run weaveworks/ignite-ubuntu --cpus 2 --memory 4GB --size 15GB --ssh=~/.ssh/myorb_bootstrap.pub --ports 5000:5000 --ports 6443:6443 --name third
sudo ignite run weaveworks/ignite-ubuntu --cpus 2 --memory 4GB --size 15GB --ssh=~/.ssh/myorb_bootstrap.pub --ports 5000:5000 --ports 6443:6443 --name fourth
```

Make sure your orb repo contains a orbiter.yml file similar to [this example](./examples/k8s/static/orbiter.yml). Show your VMs IPs with `sudo ignite ps -a`  

```bash
# Your environment is ready now, finally we can do some actual work. Launch a local orbiter that bootstraps your orb
orbctl takeoff

# When the orbiter exits, overwrite your kubeconfig by the newly created admin kubeconfig
mkdir -p ~/.kube && orbctl readsecret myorbprod_kubeconfig > ~/.kube/config

# Watch your nodes become ready
kubectl get nodes --watch

# Watch your pods become ready
kubectl get pods --all-namespaces --watch

# [Optional] Teach your ssh daemon to use the newly created ssh key for connecting to the VMS directly. The bootstrap key is not going to work anymore.
orbctl readsecret myorbprodstatic_maintenancekey > ~/.ssh/myorb-maintenance && chmod 0600 ~/.ssh/myorb-maintenance && ssh-add ~/.ssh/myorb-maintenance

# Cleanup your environment
sudo ignite rm -f $(sudo ignite ps -aq)

# Delete your git repository to clean up this tests
```

## Supported Clusters

See [Clusters](./docs/clusters.md) for details.

## Supported Providers

See [Providers](./docs/providers.md) for details.

## How To Contribute

See [contribute](./docs/contribute.md) for details

## License

The full functionality of the operator is and stays open source and free to use for everyone. We pay our wages by using Orbiter for selling further workload enterprise services like support, monitoring and forecasting, IAM, CI/CD, secrets management etc. Visit our [website](https://caos.ch) and get in touch.

See the exact licensing terms [here](./LICENSE)

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
