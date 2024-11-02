#!/bin/bash

set -euo pipefail

stop() {
    docker stop $(docker ps | grep "tini -- go" | awk '{ print $1 }') || true
}

stop
trap stop EXIT ERR SIGINT SIGTERM

docker build --progress plain -t gatekeeper .
exec docker run --rm -p 8080:8080 gatekeeper