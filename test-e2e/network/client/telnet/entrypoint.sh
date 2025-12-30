#!/bin/bash

set -uo pipefail

# if SERVER_PERMISSIONS is set, it contains the permissions to run the server
# IF not exith with 1
if [[ -z "${SERVER_PERMISSIONS:-}" ]]; then
    echo "SERVER_PERMISSIONS is not set. Exiting."
    exit 1
fi

/gatekeeper run ${SERVER_PERMISSIONS} -- sh -c "telnet 1.1.1.1 80 <<< $'GET / HTTP/1.0\r\n\r\n'" 1>/proc/1/fd/1 2>/proc/1/fd/2 
exit $?