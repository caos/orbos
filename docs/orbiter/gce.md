# Using the GCEProvider

In the following example we will create a `kubernetes` cluster on a `GCEProvider`. All the `GCEProvider` needs besides a writable Git Repository is a billable Google Cloud Project and a Google Service Account with sufficient permissions.

## Setup the Orb

To setup the configuration for an Orb, please follow [this guide](../orb/orb.md).

## Configure a billable Google Cloud Platform project of your choice

```bash
MY_GCE_PROJECT="$(gcloud config get-value project)"
ORBOS_SERVICE_ACCOUNT_NAME=orbiter-system
ORBOS_SERVICE_ACCOUNT=${ORBOS_SERVICE_ACCOUNT_NAME}@${MY_GCE_PROJECT}.iam.gserviceaccount.com

# Create a service account for the ORBITER user
gcloud iam service-accounts create ${ORBOS_SERVICE_ACCOUNT_NAME} \
    --description="${ORBOS_SERVICE_ACCOUNT_NAME}" \
    --display-name="${ORBOS_SERVICE_ACCOUNT_NAME}"

# Assign the service account the roles `Compute Admin`, `IAP-secured Tunnel User` and `Service Usage Admin`
gcloud projects add-iam-policy-binding ${MY_GCE_PROJECT} \
    --member=serviceAccount:${ORBOS_SERVICE_ACCOUNT} \
    --role=roles/compute.admin
gcloud projects add-iam-policy-binding ${MY_GCE_PROJECT} \
    --member=serviceAccount:${ORBOS_SERVICE_ACCOUNT} \
    --role=roles/iap.tunnelResourceAccessor
gcloud projects add-iam-policy-binding ${MY_GCE_PROJECT} \
    --member=serviceAccount:${ORBOS_SERVICE_ACCOUNT} \
    --role=roles/serviceusage.serviceUsageAdmin


# Create a JSON key for the service account
gcloud iam service-accounts keys create /tmp/key.json \
  --iam-account ${ORBOS_SERVICE_ACCOUNT}

# Encrypt and write the created JSON key to the orbiter.yml
orbctl writesecret orbiter.gce.jsonkey --file /tmp/key.json
rm -f /tmp/key.json
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
