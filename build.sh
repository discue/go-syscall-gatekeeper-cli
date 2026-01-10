#!/bin/bash

set -euo pipefail

export CGO_ENABLED=1

go build -o cuandari .
