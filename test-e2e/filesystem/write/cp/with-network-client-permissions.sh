#!/bin/bash

set -uo pipefail

declare -r main_path="$1"

$main_path run \
--allow-network-client \
cp run.sh .tmp/run.sh

if [[ $? -ne 0 ]]; then
    exit 0
fi

exit 1
