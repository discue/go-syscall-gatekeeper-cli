#!/bin/bash

set -uo pipefail

declare -r main_path="$1"
declare -r script_path="$( dirname -- "${BASH_SOURCE[0]}"; )";   # Get the directory name

$main_path run --allow-file-system-write node $script_path/env-var.js