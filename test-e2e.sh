#!/bin/bash

set -e

export CGO_ENABLED=1

# Cleanup function to prune Docker resources on exit (runs on success or failure)
cleanup() {
    echo "Pruning Docker containers, images, volumes, and networks (cleanup)"
    docker container prune -f || true
    docker image prune -af || true
    docker volume prune -f || true
    docker network prune -f || true
    docker system prune -af || true
}

# Ensure cleanup runs on EXIT (success, failure, or interruption)
trap cleanup EXIT

cd test-e2e && go run runner.go "$@"