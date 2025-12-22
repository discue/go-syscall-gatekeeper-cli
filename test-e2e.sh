#!/bin/bash

set -e

export CGO_ENABLED=1

cd test-e2e && go run runner.go $@