SSHKEYFILE=$1
IP=$2
PROFILE=$3
OUTPUT=$4

TMPFILENAME=tmp_profile

ssh -i ${SSHKEYFILE} orbiter@${IP} "wget http://localhost:6060/debug/pprof/${PROFILE} -O ${TMPFILENAME}"

scp -i ${SSHKEYFILE} orbiter@${IP}:/home/orbiter/${TMPFILENAME} ${OUTPUT}

ssh -i ${SSHKEYFILE} orbiter@${IP} "rm -f ${TMPFILENAME}" ${INSTANCE}