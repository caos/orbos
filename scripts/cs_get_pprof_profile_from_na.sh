ORBFILE=$1
INSTANCE=$2
IP=$3
PROFILE=$4
OUTPUT=$5

TMPFILENAME=tmp_profile

orbctl -f ${ORBFILE} node exec --command="wget http://localhost:6060/debug/pprof/${PROFILE} -O ${TMPFILENAME}" ${INSTANCE}

scp orbiter@${IP}:/home/orbiter/${TMPFILENAME} ${OUTPUT}

orbctl -f ${ORBFILE} node exec --command="rm -f ${TMPFILENAME}" ${INSTANCE}