#!/bin/bash

set -uo pipefail

declare -r main_path="$1"

$main_path

if [[ $? -ne 100 ]]; then
    exit 1
fi

exit 0