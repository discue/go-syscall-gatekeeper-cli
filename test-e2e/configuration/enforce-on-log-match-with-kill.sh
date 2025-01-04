#!/bin/bash

set -xuo pipefail

declare -r main_path="$1"
declare -r script_path="$( dirname -- "${BASH_SOURCE[0]}"; )";   # Get the directory name

$main_path run \
--allow-file-system-read \
--no-enforce-on-startup \
--trigger-enforce-on-log-match="enforce gatekeeping" \
--on-syscall-denied=error \
bash -c $script_path/log-match.bash

# This is the list of exit codes for wget:
# 4       Network failure

if [[ $? -ne 4   ]] || [[ $? -eq 159 ]]; then
    exit 1
fi

exit 0