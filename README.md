# ORBOS - GitOps everything

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

Copy the files [orbiter.yml](examples/orbiter/gce/orbiter.yml) and [boom.yml](examples/boom/boom.yml) to the root of a new git Repository.

### Configure your local environment

```bash
# Install the latest orbctl
curl -s https://api.github.com/repos/caos/orbos/releases/latest | grep "browser_download_url.*orbctl-$(uname)-$(uname -m)" | cut -d '"' -f 4 | sudo wget -i - -O /usr/local/bin/orbctl
sudo chmod +x /usr/local/bin/orbctl
sudo chown $(id -u):$(id -g) /usr/local/bin/orbctl

# Create an orb file at ${HOME}/.orb/config
orbctl configure --repourl git@github.com:me/my-orb.git --masterkey "$(openssl rand -base64 21)"
```

### Configure a billable Google Cloud Platform project of your choice

```bash
MY_GCE_PROJECT="$(gcloud config get-value project)"
ORBOS_SERVICE_ACCOUNT_NAME=orbiter-system

# Create a service account for the ORBITER user
gcloud iam service-accounts create ${ORBOS_SERVICE_ACCOUNT_NAME} \
    --description="${ORBOS_SERVICE_ACCOUNT_NAME}" \
    --display-name="${ORBOS_SERVICE_ACCOUNT_NAME}"

ORBOS_SERVICE_ACCOUNT=${ORBOS_SERVICE_ACCOUNT_NAME}@${MY_GCE_PROJECT}.iam.gserviceaccount.com

# Assign the service account the roles `Compute Admin`, `IAP-secured Tunnel User` and `Service Usage Admin`
gcloud projects add-iam-policy-binding ${MY_GCE_PROJECT} \
    --member=serviceAccount:${ORBOS_SERVICE_ACCOUNT} \
    --role=roles/compute.admin \
    --role=roles/iap.tunnelResourceAccessor \
    --role=roles/serviceusage.serviceUsageAdmin

# Create a JSON key for the service account
gcloud iam service-accounts keys create /tmp/key.json \
  --iam-account ${ORBOS_SERVICE_ACCOUNT}

# Encrypt and write the created JSON key to the orbiter.yml
orbctl writesecret orbiter.gce.jsonkey --file /tmp/key.json
rm -f /tmp/key.json
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
