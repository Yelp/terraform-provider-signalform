#!/bin/bash

project=$1
version=$2
iteration=$3
tf_version=$4


go get ${project}
[[ -d /dist ]] || mkdir /dist
cd /dist
fpm --deb-no-default-config-files -s dir -t deb --name ${project}-${tf_version} \
    --iteration ${iteration} --version ${version} \
    /go/bin/${project}=/nail/opt/terraform-${tf_version}/bin/
