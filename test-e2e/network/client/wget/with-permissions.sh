#!/bin/bash

set -uo pipefail

declare -r main_path="$1"

go run $main_path run --allow-file-system-read --allow-network-client wget -O - google.com
