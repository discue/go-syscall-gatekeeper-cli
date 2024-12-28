#!/bin/bash

set -uo pipefail

declare -r main_path="$1"

go run $main_path run --allow-file-system-write --no-implicit-allow cp run.sh .tmp/run.sh

if [[ $? -ne 0 ]]; then
    exit 0
fi

exit 1
exit 1
