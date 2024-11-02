#!/bin/bash

set -euo pipefail

# static vars
export KEY_VALUE_SEPARATOR=";"

# static for each deployment
export PROJECT=${GCLOUD_PROJECT:-discue-io-dev}
export REGION=${GCLOUD_REGION:-europe-west3}

# function specific
export CONCURRENCY=${CONCURRENCY:-250}
export CPU=${CPU:-1}
export IMAGE='europe-west3-docker.pkg.dev/discue-io-dev/functions/syscall-gatekeeper'
export MAX_INSTANCES=1
export MIN_INSTANCES=0
export MEMORY='512Mi'
export NETWORK=${NETWORK:-network-for-data-services}
export NETWORK_TAGS=${NETWORK_TAGS:-api,service}
export STAGE=${STAGE:-test}
export SECRETS=""
export SERVICE_ACCOUNT=${SERVICE_ACCOUNT:-cloud-run-functions}
export STAGE=${STAGE:-test}

# build new image first
gcloud builds submit --tag=europe-west3-docker.pkg.dev/discue-io-dev/functions/syscall-gatekeeper --quiet

FUNCTION_NAME="syscall-gatekeeper" \
NETWORK="network-for-untrusted-services" \
NETWORK_TAGS="untrusted" \
SERVICE_ACCOUNT="message-transformer-functions" \
SECRETS="" \
./deploy-gcp-functions.sh ""
