#!/bin/bash

set -eu


pkg_file=$1
tf_path=$2

dpkg -i "$pkg_file"
test -x "${tf_path}"/bin/terraform-provider-signalform
