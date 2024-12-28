#!/bin/bash

set -uo pipefail

declare -r main_path="$1"

$main_path run --allow-file-system-read --allow-network-client curl 123112311.com

if [[ $? -ne 0 ]]; then
    exit 0
fi

exit 1