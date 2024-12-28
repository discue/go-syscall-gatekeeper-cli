#!/bin/bash

set -uo pipefail

declare -r main_path="$1"

$main_path run --allow-file-system --no-implicit-allow touch .tmp/test.txt

if [[ $? -ne 0 ]]; then
    exit 0
fi

exit 1
