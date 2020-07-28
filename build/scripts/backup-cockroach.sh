#!/usr/bin/env bash
env=$1
bucket=$2
db=$3
folder=$4
filenamelocal=${env}-zitadel-${db}.sql
filenamebucket=${env}-zitadel-${db}-$(date +"%Y%m%d-%H-%M-%S").sql

/cockroach/cockroach.sh dump --certs-dir=/certificates --host=cockroachdb-public:26257 ${db} > ${filenamelocal}
curl -H "$(oauth2l header --json sa.json cloud-platform)" -H "Content-Type: application/json" -X POST --data-binary @${filenamelocal} "https://storage.googleapis.com/upload/storage/v1/b/${bucket}/o?uploadType=media&name=${filenamebucket}" > backup.sh
chmod +x backup.sh