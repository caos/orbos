# Using the GCEProvider

In the following example we will create a `kubernetes` cluster on a `GCEProvider`. All the `GCEProvider` needs besides a writable Git Repository is a billable Google Cloud Project and a Google Service Account with sufficient permissions.

## Initialize A Git Repository

Generate a new Deploy Key
```bash
mkdir -p ~/.ssh && ssh-keygen -t rsa -b 4096 -C "ORBOS repo key" -P "" -f /tmp/myorb_repo -q
```

Create a new Git Repository

Add the public part of your new SSH key pair to the git repositories trusted deploy keys with write access.

```
cat /tmp/myorb_repo.pub
```

Copy the files [orbiter.yml](../../examples/orbiter/gce/orbiter.yml) and [boom.yml](../../examples/boom/boom.yml) to the root of your Repository.

## Configure your local environment

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

## Enable IAP in a billable GCP project of you choice and create a service account

Please follow [this](https://cloud.google.com/iap/docs/enabling-compute-howto#iap-enable) instructions and enable IAP for the project in which the compute-instances should run.
Under this [link](https://console.cloud.google.com/apis/library/iap.googleapis.com) in the console it should be possible to activate.

Assign the service account the roles `Compute Admin`, `IAP-secured Tunnel User` and `Service Usage Admin`

Create a JSON key for the service account

Encrypt and write the created JSON key to the orbiter.yml

```bash
orbctl writesecret orbiter.gce.jsonkey --file ~/Downloads/<YOUR_JSON_KEY_FILE>
```

## Bootstrap your Kubernetes cluster on GCE

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
