# Using the Zitadel-Operator

## Setup the Orb

This step is only necessary if no already done for Orbiter or Boom.
To setup the configuration for an Orb, please follow [this guide](../orb/orb.md).

## Configure Zitadel

To configure `Zitadel`, a file with the name `zitadel.yml` has to be existent in the root directory of the Git repository.
The [example](../../examples/zitadel/zitadel.yml) can be used as basis, and so has to be copied to the root of the Git repository. 

## Configure Service-Account for Tracing

```bash
MY_GCE_PROJECT="$(gcloud config get-value project)"
ORBOS_SERVICE_ACCOUNT_NAME=zitadel-tracing
ORBOS_SERVICE_ACCOUNT=${ORBOS_SERVICE_ACCOUNT_NAME}@${MY_GCE_PROJECT}.iam.gserviceaccount.com

# Create a service account for the ZITADEL tracing 
gcloud iam service-accounts create ${ORBOS_SERVICE_ACCOUNT_NAME} \
    --description="${ORBOS_SERVICE_ACCOUNT_NAME}" \
    --display-name="${ORBOS_SERVICE_ACCOUNT_NAME}"

# Assign the service account the roles `Cloud Trace Agent`
gcloud projects add-iam-policy-binding ${MY_GCE_PROJECT} \
    --member=serviceAccount:${ORBOS_SERVICE_ACCOUNT} \
    --role=roles/cloudtrace.agent

# Create a JSON key for the service account
gcloud iam service-accounts keys create /tmp/key.json \
  --iam-account ${ORBOS_SERVICE_ACCOUNT}

# Encrypt and write the created JSON key to the orbiter.yml
orbctl writesecret zitadel.tracingserviceaccountjson --file /tmp/key.json
rm -f /tmp/key.json
```

## Create Bucket and configure Service-Account for Backups into Google Cloud Storage Bucket

!!!The ID used in the `zitadel.yml` for the backup-kind is used to write the secret, which means `zitadel.bucket.serviceaccountjson` means the ID is `bucket`!!!

```bash
MY_GCE_PROJECT="$(gcloud config get-value project)"
ORBOS_SERVICE_ACCOUNT_NAME=zitadel-backups
ORBOS_SERVICE_ACCOUNT=${ORBOS_SERVICE_ACCOUNT_NAME}@${MY_GCE_PROJECT}.iam.gserviceaccount.com

# Create a service account for the ZITADEL tracing 
gcloud iam service-accounts create ${ORBOS_SERVICE_ACCOUNT_NAME} \
    --description="${ORBOS_SERVICE_ACCOUNT_NAME}" \
    --display-name="${ORBOS_SERVICE_ACCOUNT_NAME}"

# Assign the service account the roles `Storage Object Admin`
gcloud projects add-iam-policy-binding ${MY_GCE_PROJECT} \
    --member=serviceAccount:${ORBOS_SERVICE_ACCOUNT} \
    --role=roles/storage.objectAdmin

# Create a JSON key for the service account
gcloud iam service-accounts keys create /tmp/key.json \
  --iam-account ${ORBOS_SERVICE_ACCOUNT}

# Encrypt and write the created JSON key to the orbiter.yml
orbctl writesecret zitadel.bucket.serviceaccountjson --file /tmp/key.json
rm -f /tmp/key.json
```

## Create and configure APIToken for Cloudflare

!!!Unfortunately it is currently not possible to create API tokens which are not bound to an existing user on Cloudflare!!!

Create an API token with enough permissions for the desired domain on Cloudflare, [here](https://developers.cloudflare.com/api/tokens/create) the official documentation.
It's necessary to create, edit and delete DNS entries and firewall rules.

```bash
orbctl writesecret zitadel.credentials.user --value ${CLOUDFLAREUSER}
orbctl writesecret zitadel.credentials.apikey --value ${APIKEY}
```

Get the `Origin CA Key` of your user under My Profile -> API Tokens -> API Keys -> Origin CA Key -> View

```bash
orbctl writesecret zitadel.credentials.userservicekey --value ${ORIGINCAKEY}
```

## Example Basic Configuration for Zitadel

Set the keys which are used for verification internally by `Zitadel`

```bash
echo 'OTPVerificationKey_1: passphrasewhichneedstobe32bytes!
UserVerificationKey_1: passphrasewhichneedstobe32bytes!
OIDCKey_1: passphrasewhichneedstobe32bytes!
CookieKey_1: passphrasewhichneedstobe32bytes!
DomainVerificationKey_1: passphrasewhichneedstobe32bytes!
IdpConfigVerificationKey_1: passphrasewhichneedstobe32bytes!' > /tmp/keys
orbctl writesecret zitadel.keys --file /tmp/keys
rm -f /tmp/keys
```

Set key used to send emails

```bash
orbctl writesecret zitadel.emailappkey --value "${EMAILAPPKEY}"
```

Set Google Chat webhook to send notifications to, [Google Chat](https://developers.google.com/hangouts/chat/how-tos/webhooks)

```bash
orbctl writesecret zitadel.googlechaturl --value "${WEBHOOKURL}"
```

Set SID and AuthToken for Twilio

```bash
orbctl writesecret zitadel.twiliosid --value "${TWILIOSID}"
orbctl writesecret zitadel.twilioauthtoken --value "${TWILIOAUTHTOKEN}"
```

## Deploy Zitadel-Operator on Orb

If you have a Kubernetes bootstraped with Orbos:

```bash
orbctl takeoff
```

If you want to deploy Zitadel-Operator to a Kubernetes boostraped otherwise:

```bash
orbctl takeoff --kubeconfig ${KUBECONFIG}
```


