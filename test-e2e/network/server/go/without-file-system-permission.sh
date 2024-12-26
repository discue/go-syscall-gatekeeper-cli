#!/bin/bash

set -xuo pipefail

declare -r main_path="$1"

# Start the server in the background
nohup go run $main_path run --allow-network-server ./.tmp/server > /dev/null 2>&1 &

server_pid=$!
trap 'kill -9 $server_pid' EXIT

# Number of retries
max_retries=5

# Wait for the server to start and check the status code.
for i in $(seq 1 $max_retries); do
    status_code=$(curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8081)
    
    if [[ "$status_code" == "200" ]]; then # Changed from "40" to "404"
        echo "Server returned 200 as expected."
        exit 1
        elif [[ "$status_code" == "" ]]; then
        echo "Server not up yet. Retrying..."
    fi
    
    sleep 1
done

exit 0