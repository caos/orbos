env=$1
bucket=$2
db=$3
folder=$4
safile=$5
certs=$6
timestamp=$7

filenamelocal=zitadel-${db}.sql
filenamebucket=zitadel-${db}-${timestamp}.sql

/cockroach/cockroach.sh dump --certs-dir=${certs} --host=cockroachdb-public:26257 ${db} > ${folder}/${filenamelocal}
curl -H "$(oauth2l header --json ${safile} cloud-platform)" -H "Content-Type: application/json" -X POST --data-binary @${folder}/${filenamelocal} "https://storage.googleapis.com/upload/storage/v1/b/${bucket}/o?uploadType=media&name=${env}/${timestamp}/${filenamebucket}"