#!/bin/bash

set -eu

dpkg -i "$1"
test -x /nail/opt/terraform-$2/bin/terraform-provider-signalform
