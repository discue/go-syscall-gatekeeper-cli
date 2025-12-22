#!/bin/bash

set -euo pipefail

export CGO_ENABLED=1

go test ./... -cover
