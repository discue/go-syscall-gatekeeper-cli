#!/bin/bash

set -uo pipefail

# if SERVER_PERMISSIONS is set, it contains the permissions to run the server
# IF not exith with 1
if [[ -z "${SERVER_PERMISSIONS:-}" ]]; then
    echo "SERVER_PERMISSIONS is not set. Exiting."
    exit 1
fi

echo more | /gatekeeper run ${SERVER_PERMISSIONS} -- tee -a /tmp/append-target 

# check if the append was successful
if grep -q "more" /tmp/append-target; then
    echo "Append succeeded."
    exit 0
else
    echo "Append failed."
    exit 1
fi