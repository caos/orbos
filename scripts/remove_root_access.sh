#!/bin/bash

set -ex

echo "PermitRootLogin no" >> /etc/ssh/sshd_config
service sshd restart
rm -rf /root/.ssh
