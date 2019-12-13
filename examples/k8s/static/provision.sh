#!/bin/bash

adduser orbiter
cp -r ~/.ssh /home/orbiter/.ssh
chown -R 1000:1000 /home/orbiter/.ssh
chmod 600 /etc/sudoers
echo "orbiter	ALL=(ALL)	NOPASSWD: ALL" >> /etc/sudoers