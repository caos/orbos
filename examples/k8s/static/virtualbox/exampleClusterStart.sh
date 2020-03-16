#!/usr/bin/env bash
publicKeyFile=$1
tmpFolder=$2

imageURL=http://mirror.init7.net/centos/7.7.1908/isos/x86_64/CentOS-7-x86_64-Minimal-1908.iso
imageName=centos.iso

./downloadImage.sh ${tmpFolder} ${imageName} ${imageURL}

for MACHINE in  master1 worker1
do
    ./machine.sh $MACHINE ks.cfg ${publicKeyFile} ${tmpFolder} ${imageName}
done
