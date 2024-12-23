#!/bin/bash

set -uo pipefail

declare -r main_path="$1"

touch .tmp/test.txt

go run $main_path run \
--allow-file-system-read \
--allow-network-client  \
--no-implicit-allow \
wget -O -.tmp google.com

if [[ $? -ne 0 ]]; then
    exit 0
fi

exit 1
