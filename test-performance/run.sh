#!/bin/bash

set -euo pipefail

MSYS_NO_PATHCONV=1 docker run --rm \
-i \
-p 5665:5665 \
--mount type=bind,src=.,dst=/etc/test \
ghcr.io/grafana/xk6-dashboard:latest \
run \
--out web-dashboard=export=/etc/test/report.html \
--out json=/etc/test/report.json - < api-performance.k6.js