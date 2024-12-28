#!/bin/bash

set -euo pipefail

dir=$(dirname "$0")  # Get directory of the script (possibly a symlink)

go build -o .tmp/concurrent-file-reads $dir/concurrent-file-reads.go