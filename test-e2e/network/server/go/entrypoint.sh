#!/bin/bash

set -uo pipefail

# if SERVER_PERMISSIONS is set, it contains the permissions to run the server
# IF not exith with 1
if [[ -z "${SERVER_PERMISSIONS:-}" ]]; then
    echo "SERVER_PERMISSIONS is not set. Exiting."
    exit 1
fi

/gatekeeper run ${SERVER_PERMISSIONS} -- /server 1>/proc/1/fd/1 2>/proc/1/fd/2 &

# Number of retries
max_retries=5

# Wait for the server to start and check the status code.
for i in $(seq 1 $max_retries); do
    status_code=$(curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8082)
    
    if [[ "$status_code" == "404" ]]; then
        echo "Server returned 404 as expected."
        exit 0
        elif [[ -z "$status_code" || "$status_code" == "000" ]]; then
        echo "Server not up yet. Retrying..."
    fi
    
    sleep 1
done

echo "Server did not become ready in time."
exit 1