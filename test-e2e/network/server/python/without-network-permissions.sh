#!/bin/bash

set -uo pipefail

declare -r main_path="$1"
declare -r script_path="$( dirname -- "${BASH_SOURCE[0]}"; )";   # Get the directory name
declare -r start_command="python3 -m http.server 8080"

nohup $main_path run --allow-file-system-read $start_command > /dev/null 2>&1 &

server_pid=$!
trap 'pkill -f "'"$start_command"'"' EXIT  # Corrected trap

# Number of retries
max_retries=5

# Wait for the server to start and check the status code.
for i in $(seq 1 $max_retries); do
    status_code=$(curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080)
    
    if [[ "$status_code" == "200" ]]; then # Changed from "40" to "404"
        echo "Server returned 200 as expected."
        exit 1
        elif [[ "$status_code" == "" ]]; then
        echo "Server not up yet. Retrying..."
    fi
    
    sleep 1
done

exit 0