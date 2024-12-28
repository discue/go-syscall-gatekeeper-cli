#!/bin/bash

set -uo pipefail

declare -r main_path="$1"
declare -r dir=$(dirname "$0")  # Get directory of the script (possibly a symlink)

$main_path run --allow-file-system-read .tmp/concurrent-file-reads