#!/bin/bash

set -uo pipefail

declare -r main_path="$1"

touch .tmp/test.txt

go run $main_path run \
--allow-file-system \
--allow-process-management \
--allow-memory-management \
--allow-process-synchronization \
--allow-misc \
wget google.com

if [[ $? -ne 0 ]]; then
    exit 0
fi

exit 1
