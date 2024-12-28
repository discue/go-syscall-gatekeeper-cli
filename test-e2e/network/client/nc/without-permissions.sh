#!/bin/bash

set -uo pipefail

declare -r main_path="$1"

$main_path run --allow-file-system-read --no-implicit-allow grep "done" run.sh

if [[ $? -ne 0 ]]; then
    exit 0
fi

exit 1
--allow-misc \
nc -w 1 example.com 80 << EOF
GET / HTTP/1.1
Host: example.com

EOF

if [[ $? -ne 0 ]]; then
    exit 0
fi

exit 1
