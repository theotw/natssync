#!/usr/bin/env bash

#
# Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
#

NATS_CLOUD_SERVER="${NATS_CLOUD_PORT:-nats://localhost:4222}"
NATS_ONPREM_SERVER="${NATSSYNC_CLIENT_PORT:-nats://localhost:4223}"
EXIT_APP_TOPIC="${EXIT_APP_TOPIC:-natssync.testing.exitapp}"

if [ "$(command -v gtimeout)" ]; then
  alias timeout=gtimeout
elif [ ! "$(command -v timeout)" ]; then
  echo "This script requires a timeout command (timeout or gtimeout (mac via homebrew's coreutils))"
  exit 1
fi

# Send exit signal via NATS
if [ "$(command -v nats)" ]; then
  nats pub --server="${NATS_CLOUD_SERVER}" "${EXIT_APP_TOPIC}" '' || true
  nats pub --server="${NATS_ONPREM_SERVER}" "${EXIT_APP_TOPIC}" '' || true
else
  go run apps/natstool.go -u "${NATS_CLOUD_SERVER}" -s "${EXIT_APP_TOPIC}" -m 'hello world' || true
  go run apps/natstool.go -u "${NATS_ONPREM_SERVER}" -s "${EXIT_APP_TOPIC}" -m 'hello world' || true
fi

# Wait until the relevant containers are stopped
source "$(dirname $0)/../single_cluster_test/docker-containers.sh"

for container in ${SYNCCLIENT} ${SYNCSERVER} ${HTTPPROXYLET} ${HTTPPROXYSERVER}; do
  echo "Waiting for container to stop: ${container}"
  timeout 40 docker wait "${container}" || true
done

echo "Done! All relevant containers have completed gracefully."
