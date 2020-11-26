#!/bin/bash

set -e

ORB="${1}"
KEY="${2}"

for IP in $(orbctl -f ${ORB} nodes list --column external); do
    if ssh -o StrictHostKeyChecking=no -i ${KEY} orbiter@${IP} sudo whoami
    then
	    echo "orbiter user is already done"
    else
            ssh -o StrictHostKeyChecking=no -i ${KEY} root@${IP} 'bash -s' < ./scripts/add_orbiter_user.sh
    fi

    if ssh -o StrictHostKeyChecking=no -i ${KEY} root@${IP} sudo whoami
    then
            ssh -o StrictHostKeyChecking=no -i ${KEY} orbiter@${IP} 'bash -s' < ./scripts/remove_root_access.sh
    else
	    echo "root access is already disabled"
    fi
done

