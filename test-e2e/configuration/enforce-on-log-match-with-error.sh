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

# https://specificlanguages.com/posts/2022-07/14-exit-code-137/
# Exit code 137 is Linux-specific and means that your process was killed by a signal, namely SIGKILL.
# The main reason for a process getting killed by SIGKILL on Linux (unless you do it yourself) is running out of memory.1

if [[ $? -ne 137 ]]; then
    exit 1
fi

exit 0