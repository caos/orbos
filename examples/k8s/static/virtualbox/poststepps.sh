#!/usr/bin/env bash

machineName=$1
tmpFolder=$2

rm -rf ${tmpFolder}/images
rm -rf ${tmpFolder}/kickstart

controllerNameDisk="SATAController"
controllerNameImage="IDEController"

VBoxManage storageattach ${machineName} --storagectl ${controllerNameDisk} --port 1 --device 0 --type hdd --medium none
VBoxManage storageattach ${machineName} --storagectl ${controllerNameImage} --port 0 --device 0 --type dvddrive --medium none
VBoxManage storagectl ${machineName} --name ${controllerNameImage} --remove
VBoxManage modifyvm ${machineName} --nictype1 virtio

convertedKickstartDisk="${tmpFolder}/vms/${machineName}/OEMDRV.vmdk"
kickstartDisk="${tmpFolder}/vms/${machineName}/OEMDRV.dmg"

rm ${convertedKickstartDisk}
rm ${kickstartDisk}
