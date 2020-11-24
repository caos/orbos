#!/bin/bash

adduser orbiter
usermod -aG wheel orbiter
echo "%wheel	ALL=(ALL)	NOPASSWD: ALL" >> /etc/sudoers
cp -r /root/.ssh /home/orbiter/.ssh
chown -R orbiter /home/orbiter /usr/local/bin /var/orbiter

