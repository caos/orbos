#!/usr/bin/env bash

machineName=$1
kickstartFile=$2
publicKeyPath=$3
tmpFolder=$4
image=$5

mkdir -p ${tmpFolder}

baseFolder=${tmpFolder}/vms
mkdir -p ${baseFolder}

########################################################################################################################
# VM
########################################################################################################################
VBoxManage createvm --name ${machineName} --ostype "RedHat_64" --register --basefolder ${baseFolder}
VBoxManage modifyvm ${machineName} \
    --cpus 2 \
    --ioapic on \
    --memory 2048 \
    --vram 20 \
    --nic1 bridged \
    --nictype1 82540EM \
    --audioout off \
    --graphicscontroller vmsvga \
    --bridgeadapter1 en0

machineFolder=${baseFolder}/${machineName}
diskPath=${machineFolder}/DISK.vdi
controllerNameDisk="SATAController"
controllerNameImage="IDEController"
diskSize=8000

########################################################################################################################
# os disk
########################################################################################################################
VBoxManage createmedium --filename ${diskPath} --size ${diskSize} --format VDI
VBoxManage storagectl ${machineName} \
    --name "${controllerNameDisk}" \
    --add sata \
    --controller IntelAHCI
VBoxManage storageattach ${machineName} \
    --storagectl "${controllerNameDisk}" \
    --port 0 \
    --device 0 \
    --type hdd \
    --medium ${diskPath}

########################################################################################################################
# kickstart volume
########################################################################################################################

diskFolder="${tmpFolder}/kickstart"
kickstartFilePath="${diskFolder}/ks.cfg"

mkdir -p ${diskFolder}
ORBITER_BOOTSTRAP_PUBLICKEY="$(cat ${publicKeyPath})" envsubst < ${kickstartFile} > ${kickstartFilePath}

kickstartDisk="${machineFolder}/OEMDRV.dmg"
convertedKickstartDisk="${machineFolder}/OEMDRV.vmdk"

hdiutil create -megabytes 4 -format UDIF -fs MS-DOS -volname OEMDRV  -srcfolder ${diskFolder} -ov ${kickstartDisk}
VBoxManage internalcommands createrawvmdk -filename ${convertedKickstartDisk} -rawdisk ${kickstartDisk}
#rm ${kickstartDisk}
VBoxManage storageattach ${machineName} \
    --storagectl "${controllerNameDisk}" \
    --port 1 \
    --device 0 \
    --type hdd \
    --medium "${convertedKickstartDisk}" \
    --hotpluggable on

########################################################################################################################
# installation image
########################################################################################################################
imagePath=${tmpFolder}/images/${image}
VBoxManage storagectl ${machineName} \
    --name "${controllerNameImage}" \
    --add ide \
    --controller PIIX4
VBoxManage storageattach ${machineName} \
    --storagectl "${controllerNameImage}" \
    --port 0 \
    --device 0 \
    --type dvddrive \
    --medium "${imagePath}"

VBoxManage modifyvm ${machineName} \
    --boot1 dvd \
    --boot2 disk \
    --boot3 none \
    --boot4 none
