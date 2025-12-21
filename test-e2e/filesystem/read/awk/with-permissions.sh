#!/bin/bash

set -xuo pipefail

declare -r main_path="$1"

$main_path run \
--allow-file-system-read \
--allow-process-management \
--allow-memory-management \
--allow-process-synchronization \
--allow-misc \
-- \
awk '{ print }' run.sh
