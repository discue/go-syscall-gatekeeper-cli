#!/bin/bash

set -xuo pipefail

# If SERVER_PERMISSIONS is set, it contains the flags to gatekeep the server.
if [[ -z "${SERVER_PERMISSIONS:-}" ]]; then
    echo "SERVER_PERMISSIONS is not set. Exiting."
    exit 1
fi

# Start the Node server under gatekeeper enforcement (run mode).
/gatekeeper run ${SERVER_PERMISSIONS} -- node /server.cjs > /dev/null 2>&1 &

declare -r gatekeeper_pid=$!

# Number of retries
max_retries=5
success=0

# Wait for the server to start and check the status code.
for i in $(seq 1 $max_retries); do
    status_code=$(curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080)
    
    if [[ "$status_code" == "200" ]]; then # Changed from "40" to "404"
        echo "Server returned 200 as expected."
        success=1
        break
        elif [[ "$status_code" == "" ]]; then
        echo "Server not up yet. Retrying..."
    fi
    
    sleep 1
done

# If no success until here exit early
if [[ $success == 0 ]]; then
    exit 1
fi

# Now stop the server
kill -TERM ${gatekeeper_pid}

# And expect no status 200 anymore because server was stopped
for i in $(seq 1 $max_retries); do
    status_code=$(curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080)
    
    if [[ "$status_code" == "200" ]]; then # Changed from "40" to "404"
        echo "Server returned 200 but we didn't expect this time."
        exit 1
        elif [[ "$status_code" == "" ]]; then
        break
    fi
    
    sleep 1
done

exit 0