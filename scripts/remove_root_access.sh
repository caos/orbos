#!/bin/bash

set -ex

echo "PermitRootLogin no" | sudo tee -a /etc/ssh/sshd_config
sudo service sshd restart
sudo rm -rf /root/.ssh
