#!/bin/bash

set -uo pipefail

declare -r main_path="$1"

$main_path run --allow-file-system-read --allow-network-client curl google.com
