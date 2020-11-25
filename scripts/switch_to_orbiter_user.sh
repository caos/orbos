#!/bin/bash

set -ex

ORB="${1}"
KEY="${2}"

orbctl -f ${ORB} nodes list --column external | while read IP; do
    ssh -o StrictHostKeyChecking=no -i ${KEY} root@${IP} 'bash -s' < ./scripts/add_orbiter_user.sh
    ssh -i ${KEY} orbiter@${IP} 'bash -s' < ./scripts/remove_root_access.sh
done

