#!/usr/bin/env bash

KUBECTL="kubectl --kubeconfig /Users/christianjakob/.kube/zitadel/dev"
SED="/usr/local/bin/gsed"
DOCKER_LS="/Users/christianjakob/caos/internal/docker-versions/docker-ls"



function getImages {
images=$($KUBECTL get pods --all-namespaces -o jsonpath="{.items[*].spec.containers[*].image}" |\
tr -s '[[:space:]]' '\n' |\
sort |\
uniq)
}


function requestRemoteVersions() {
DOCKERHUB_IMAGE=$1
DOCKER_REPO=$2

if [[ $2 == "" ]];
  then "$DOCKER_LS" tags "$DOCKERHUB_IMAGE" | sort --version-sort | tail -n 15 ;
  else "$DOCKER_LS" -r "https://""$DOCKER_REPO" tags "$DOCKERHUB_IMAGE" | sort --version-sort | tail -n 15 ;
fi
}


function isolatePackagesandVersion(){
    local image version ISURL
    for fields in ${images[@]}
    do
            IFS=$':' read -r image version <<< "$fields"
            echo "image: $image version: $version "  
            # simple check for an URL domain name to use custom repo for quay.io, ghcr.io etc
            if [[ "$image" =~  ^(.*\.[a-zA-Z]*)\/(.*)$ ]] ;
              then requestRemoteVersions "${BASH_REMATCH[2]}" "${BASH_REMATCH[1]}"; 
            # second check for uniq name, library/ needs to be prefixed then
            elif [[ "$image" =~  ^(.*\/.*)$ ]] ;
            then  requestRemoteVersions $image ;
            #else put library in front of it 
            else requestRemoteVersions "library/"$image ; 
            fi
    done
}


getImages
isolatePackagesandVersion
