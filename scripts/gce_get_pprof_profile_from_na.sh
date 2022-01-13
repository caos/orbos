INSTANCE=orbiter@$1
PROFILE=$2
OUTPUT=$3

TMPFILENAME=tmp_profile

gcloud compute ssh ${INSTANCE} --command="wget http://localhost:6060/debug/pprof/${PROFILE} -O ${TMPFILENAME}"

gcloud compute scp ${INSTANCE}:/home/orbiter/${TMPFILENAME} ${OUTPUT}

gcloud compute ssh ${INSTANCE} --command="rm -f ${TMPFILENAME}"