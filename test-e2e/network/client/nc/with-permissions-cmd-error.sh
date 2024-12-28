#!/bin/bash

set -uo pipefail

declare -r main_path="$1"

$main_path run --allow-file-system-read --allow-network-client nc -w 1 example.com 80 << EOF
GET / HTTP/1.1
Host: 123112311.com
Connection: close

EOF