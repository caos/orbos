#!/bin/bash

set -ex

echo "PermitRootLogin no" | sudo tee -a /etc/ssh/sshd_config
sudo rm -rf /root/.ssh
sudo service sshd restart
