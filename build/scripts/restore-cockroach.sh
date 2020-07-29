bucket=$1
env=$2
timestamp=$3
db=$4
safile=$5
certs=$6

urlencode() {
    # urlencode <string>
    old_lc_collate=$LC_COLLATE
    LC_COLLATE=C

    local length="${#1}"
    for (( i = 0; i < length; i++ )); do
        local c="${1:i:1}"
        case $c in
            [a-zA-Z0-9.~_-]) printf "$c" ;;
            *) printf '%%%02X' "'$c" ;;
        esac
    done

    LC_COLLATE=$old_lc_collate
}

filenamelocal=zitadel-${db}.sql
filenamebucket=zitadel-${db}-${timestamp}.sql

curl -X GET \
  -H "$(oauth2l header --json ${safile} cloud-platform)" \
  -o "${filenamelocal}" \
  "https://storage.googleapis.com/storage/v1/b/${bucket}/o/$(urlencode ${env}/${timestamp}/${filenamebucket})?alt=media"

/cockroach/cockroach.sh sql --certs-dir=${certs} --host=cockroachdb-public:26257 -e "DROP DATABASE ${db};"
/cockroach/cockroach.sh sql --certs-dir=${certs} --host=cockroachdb-public:26257 -e "CREATE DATABASE ${db};"
/cockroach/cockroach.sh sql --certs-dir=${certs} --host=cockroachdb-public:26257 --database=${db} < ${filenamelocal}

