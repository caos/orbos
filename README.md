# Orbos - GitOps everything

![ORBOS](./docs/img/orbos-logo-oneline-lightdesign@2x.png)

[![semantic-release](https://img.shields.io/badge/%20%20%F0%9F%93%A6%F0%9F%9A%80-semantic--release-e10079.svg)](https://github.com/semantic-release/semantic-release)
[![Release](https://github.com/caos/orbos/workflows/Release/badge.svg)](https://github.com/caos/orbos/actions)
[![license](https://badgen.net/github/license/caos/orbos/)](https://github.com/caos/orbos/blob/master/LICENSE)
[![release](https://badgen.net/github/release/caos/orbos/stable)](https://github.com/caos/orbos/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/caos/orbos)](https://goreportcard.com/report/github.com/caos/orbos)
[![codecov](https://codecov.io/gh/caos/orbos/branch/master/graph/badge.svg)](https://codecov.io/gh/caos/orbos)

> This project is in alpha state. The API will continue breaking until version 1.0.0 is released

## [ORBOS explained](docs/explained.md)

### [ORBITER](docs/orbiter/orbiter.md)

### [BOOM](docs/boom/boom.md)

## Getting Started on Google Compute Engine

In the following example we will create a `kubernetes` cluster on a `GCEProvider`. All the `GCEProvider` needs besides a writable Git Repository is a billable Google Cloud Project and a Google Service Account with sufficient permissions.

### Initialize A Git Repository

Generate a new Deploy Key
```bash
mkdir -p ~/.ssh && ssh-keygen -t rsa -b 4096 -C "ORBOS repo key" -P "" -f /tmp/myorb_repo -q
```

Create a new Git Repository

Add the public part of your new SSH key pair to the git repositories trusted deploy keys with write access.

```
cat /tmp/myorb_repo.pub
```

Copy the files [orbiter.yml](examples/orbiter/gce/orbiter.yml) and [boom.yml](examples/boom/boom.yml) to the root of your Repository.

### Configure your local environment

Download the latest orbctl

```bash
curl -s https://api.github.com/repos/caos/orbos/releases/latest | grep "browser_download_url.*orbctl-$(uname)-$(uname -m)" | cut -d '"' -f 4 | sudo wget -i - -O /usr/local/bin/orbctl
sudo chmod +x /usr/local/bin/orbctl
sudo chown $(id -u):$(id -g) /usr/local/bin/orbctl
```

Create an orb file

```bash
mkdir -p ~/.orb
cat > ~/.orb/config << EOF
url: git@github.com:me/my-orb.git
masterkey: $(openssl rand -base64 21)
repokey: |
$(sed s/^/\ \ /g /tmp/myorb_repo)
EOF
```

### Create a service account in a billable GCP project of your choice

Assign the service account the roles `Compute Admin`, `IAP-secured Tunnel User` and `Service Usage Admin`

Create a JSON key for the service account

Encrypt and write the created JSON key to the orbiter.yml

```bash
orbctl writesecret orbiter.gce.jsonkey --file ~/Downloads/<YOUR_JSON_KEY_FILE>
```

### Bootstrap your Kubernetes cluster on GCE

```bash
orbctl takeoff
```

As soon as the Orbiter has deployed itself to the cluster, you can decrypt the generated admin kubeconfig

```bash
mkdir -p ~/.kube
orbctl readsecret k8s.kubeconfig > ~/.kube/config
```

Wait for grafana to become running

```bash
kubectl --namespace caos-system get po -w
```

Open your browser at localhost:8080 to show your new clusters dashboards

```bash
kubectl --namespace caos-system port-forward svc/grafana 8080:80
```

Delete everything created by Orbiter

```bash
orbctl destroy
```


## License

The full functionality of the operator is and stays open source and free to use for everyone. We pay our wages by using Orbiter for selling further workload enterprise services like support, monitoring and forecasting, IAM, CI/CD, secrets management etc. Visit our [website](https://caos.ch) and get in touch.

See the exact licensing terms [here](./LICENSE)

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
