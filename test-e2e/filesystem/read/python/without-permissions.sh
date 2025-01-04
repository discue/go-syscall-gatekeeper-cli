#!/bin/bash

set -uo pipefail

declare -r main_path="$1"
declare -r script_path="$( dirname -- "${BASH_SOURCE[0]}"; )";   # Get the directory name

$main_path run \
--allow-process-management \
--allow-memory-management \
--allow-process-synchronization \
--allow-misc \
python3 $script_path/read.py

if [[ $? -ne 0 ]]; then
    exit 0
fi

exit 1
