#!/bin/bash

set -uo pipefail

declare -r main_path="$1"

touch .tmp/test.txt

$main_path run --allow-file-system-write --no-implicit-allow mv .tmp/test.txt .tmp/copy.txt

if [[ $? -ne 0 ]]; then
    exit 0
fi

exit 1