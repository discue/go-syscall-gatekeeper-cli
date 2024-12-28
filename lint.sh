#/bin/bash

set -euo pipefail

MSYS_NO_PATHCONV=1 docker run --rm \
--mount type=bind,src=.,dst=/app \
-w /app golangci/golangci-lint golangci-lint \
run -v
