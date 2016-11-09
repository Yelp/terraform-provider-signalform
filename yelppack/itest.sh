#!/bin/bash

set -eu

dpkg -i "$1"
test -x /nail/opt/bin/terraform-provider-signalform
