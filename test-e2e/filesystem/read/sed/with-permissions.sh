#!/bin/bash

set -uo pipefail

declare -r main_path="$1"

$main_path run --allow-file-system-read sed -n '10,20p' run.sh