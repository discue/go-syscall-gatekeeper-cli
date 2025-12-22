#!/bin/bash

set -xuo pipefail

# If SERVER_PERMISSIONS is set, it contains the flags to gatekeep the server.
if [[ -z "${SERVER_PERMISSIONS:-}" ]]; then
    echo "SERVER_PERMISSIONS is not set. Exiting."
    exit 1
fi

# Start the Node server under gatekeeper enforcement (run mode).
/gatekeeper run ${SERVER_PERMISSIONS} -- python3 -m http.server 8080 > /dev/null 2>&1 &

# Number of retries
max_retries=5

# Wait for the server to start and check the status code.
for i in $(seq 1 $max_retries); do
    status_code=$(curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080)

    if [[ "$status_code" == "200" ]]; then
        echo "Server returned 200 as expected."
        exit 0
    elif [[ -z "$status_code" || "$status_code" == "000" ]]; then
        echo "Server not up yet. Retrying..."
    fi

    sleep 2
done

echo "Server did not become ready in time."
exit 1