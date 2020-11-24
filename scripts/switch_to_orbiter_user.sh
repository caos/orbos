#!/bin/bash

set -ex

KEY="${1}"
IP="${2}"

ssh -o StrictHostKeyChecking=no -i ${KEY} root@${IP} 'bash -s' < ./scripts/add_orbiter_user.sh
ssh -i ${KEY} orbiter@${IP} 'bash -s' < ./scripts/remove_root_access.sh
