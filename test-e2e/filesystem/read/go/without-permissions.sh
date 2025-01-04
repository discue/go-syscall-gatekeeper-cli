#!/bin/bash

set -uo pipefail

declare -r main_path="$1"

$main_path run \
--allow-process-management \
--allow-memory-management \
--allow-process-synchronization \
--allow-misc \
.tmp/read

if [[ $? -ne 0 ]]; then
    exit 0
fi

exit 1
