#!/bin/bash

set -uo pipefail

declare -r main_path="$1"

touch .tmp/test.txt

go run $main_path run \
--allow-file-system \
--allow-networking  \
--no-implicit-allow \
wget -P .tmp google.com

if [[ $? -ne 0 ]]; then
    exit 0
fi

exit 1
