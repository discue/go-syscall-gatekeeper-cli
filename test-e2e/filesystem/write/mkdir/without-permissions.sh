#!/bin/bash

set -uo pipefail

declare -r main_path="$1"

go run $main_path run \
--allow-process-management \
--allow-memory-management \
--allow-process-synchronization \
--allow-misc \
mkdir .tmp/dir

if [[ $? -ne 0 ]]; then
    exit 0
fi

exit 1