certs=$1
db=$2

/cockroach/cockroach.sh sql --certs-dir=${certs} --host=cockroachdb-public:26257 -e "DROP DATABASE IF EXISTS ${db} CASCADE;"
/cockroach/cockroach.sh sql --certs-dir=${certs} --host=cockroachdb-public:26257 -e "DROP USER IF EXISTS ${db};"
/cockroach/cockroach.sh sql --certs-dir=${certs} --host=cockroachdb-public:26257 -e "TRUNCATE defaultdb.flyway_history_schema;"
