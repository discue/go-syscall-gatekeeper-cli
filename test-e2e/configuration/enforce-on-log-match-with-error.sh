#!/bin/bash

set -xuo pipefail

declare -r main_path="$1"
declare -r script_path="$( dirname -- "${BASH_SOURCE[0]}"; )";   # Get the directory name

$main_path run \
--allow-file-system-read \
--no-enforce-on-startup \
--trigger-enforce-on-log-match="enforce gatekeeping" \
--on-syscall-denied=kill \
bash -c $script_path/log-match.bash

if [[ $? -ne 4 && $? -ne 159 ]]; then
    exit 1
fi

exit 0