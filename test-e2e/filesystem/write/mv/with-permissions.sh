#!/bin/bash

set -uo pipefail

declare -r main_path="$1"

touch .tmp/test.txt

go run $main_path run --allow-file-system-write mv .tmp/test.txt .tmp/copy.txt
