#!/usr/bin/env bash

tmpFolder=$1

imageURL=http://mirror.init7.net/centos/7.8.2003/isos/x86_64/CentOS-7-x86_64-Minimal-2003.iso

imageFolder=${tmpFolder}/images
mkdir -p ${imageFolder}

imagePath=${imageFolder}/centos.iso
wget -O ${imagePath} ${imageURL}