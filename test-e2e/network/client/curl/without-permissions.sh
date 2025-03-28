#!/bin/bash

set -uo pipefail

declare -r main_path="$1"

$main_path run \
--allow-file-system-read \
--allow-process-management \
--allow-memory-management \
--allow-process-synchronization \
--allow-misc \
curl google.com

if [[ $? -ne 0 ]]; then
    exit 0
fi

exit 1
