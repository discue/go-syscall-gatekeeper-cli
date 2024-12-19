#!/bin/bash

set -uo pipefail

declare -r main_path="$1"

touch .tmp/test.txt

go run $main_path run --allow-file-system wget 123112311.com

if [[ $? -ne 0 ]]; then
    exit 0
fi

exit 1
