#!/bin/bash -exv

project=$1
version=$2
iteration=$3
tf_version=$4
tf_path=$5

go get ${project}

[[ -d /dist ]] || mkdir /dist
cd /dist

fpm \
    --deb-no-default-config-files \
    -s dir \
    -t deb --name ${project}-${tf_version} \
    --iteration ${iteration} \
    --version ${version} \
    /go/bin/${project}="${tf_path}"/bin/

env GOOS=${GOOS} GOARCH=${GOARCH} go build -v \
    -o /dist/terraform-provider-signalform-${GOOS}_${GOARCH} \
    terraform-provider-signalform
