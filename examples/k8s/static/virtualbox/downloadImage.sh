#!/usr/bin/env bash

tmpFolder=$1
imageName=$2
imageURL=$3

imageFolder=${tmpFolder}/images
mkdir -p ${imageFolder}

imagePath=${imageFolder}/${imageName}
wget -O ${imagePath} ${imageURL}