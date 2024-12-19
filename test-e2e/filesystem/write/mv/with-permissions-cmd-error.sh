#!/bin/bash

set -uo pipefail

declare -r main_path="$1"

go run $main_path run --allow-file-system mv .tmp/test.txt .tmp/copy.txt

if [[ $? -ne 0 ]]; then
    exit 0
fi

exit 1