#!/bin/bash

set -e

dir=$(dirname "$0")  # Get directory of the script (possibly a symlink)

go build -o .tmp/server $dir/server.go