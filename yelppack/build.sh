#!/bin/bash

project=$1
version=$2
iteration=$3

go get ${project}
[[ -d /dist ]] || mkdir /dist
cd /dist
fpm --deb-no-default-config-files -s dir -t deb --name ${project} \
    --iteration ${iteration} --version ${version} \
    /go/bin/${project}=/nail/opt/bin/
