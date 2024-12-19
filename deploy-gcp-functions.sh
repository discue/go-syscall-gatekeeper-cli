#!/bin/bash

set -euo pipefail

declare -r KEY_VALUE_SEPARATOR=";"

declare -r FUNCTION_NAME=${FUNCTION_NAME}

declare -r CONCURRENCY=${CONCURRENCY:-100}
declare -r CPU=${CPU:-1}
declare -r IMAGE=${IMAGE}
declare -r INGRESS=${INGRESS:-all}
declare -r INVOCATION=--${ALLOW_UNAUTHENTICATED:-''}allow-unauthenticated
declare -r MAX_INSTANCES=5
declare -r MEMORY='512Mi'
declare -r MIN_INSTANCES=0
declare -r NETWORK=${NETWORK:-network-for-data-services}
declare -r NETWORK_TAGS=${NETWORK_TAGS:-api,service}
declare -r PROJECT=${GCLOUD_PROJECT:-discue-io-dev}
declare -r REGION=${GCLOUD_REGION:-europe-west3}
declare -r SECRETS=${SECRETS}
declare -r SERVICE_ACCOUNT=${SERVICE_ACCOUNT:-cloud-run-functions}
declare -r STAGE=${STAGE:-test}
declare -r TIMEOUT=${TIMEOUT:-30}

deployHttpsFunction() {
    name=$1
    
    # we update container specific configuration as well
    gcloud beta run deploy ${name} \
    --args="" \
    --command="" \
    --concurrency=${CONCURRENCY} \
    --cpu=${CPU} \
    --cpu-boost \
    --execution-environment=gen2 \
    --image=${IMAGE} \
    --ingress=${INGRESS} \
    ${INVOCATION} \
    --max-instances=${MAX_INSTANCES} \
    --min-instances=${MIN_INSTANCES} \
    --memory=${MEMORY} \
    --network=${NETWORK} \
    --network-tags=${NETWORK_TAGS} \
    --project=${PROJECT} \
    --region=${REGION} \
    --service-account=${SERVICE_ACCOUNT}@${PROJECT}.iam.gserviceaccount.com \
    --subnet=${NETWORK} \
    --no-use-http2 \
    --no-session-affinity \
    --vpc-egress=private-ranges-only
}

removeOldRevisions() {
    name=$1
    
    gcloud run revisions list \
    --filter="status.conditions.type:Active AND status.conditions.status:'False'" --sort-by deployed --region ${REGION}| \
    grep ${name} | awk '{if(NR > 5) print $2}' | \
    xargs -r -L1 gcloud run revisions delete --region ${REGION} --quiet
}

deployHttpsFunction "${FUNCTION_NAME}" "${IMAGE}"
removeOldRevisions "${FUNCTION_NAME}"