#!/usr/bin/env bash

#
# Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
#

NATS_CLOUD_SERVER="${NATS_CLOUD_PORT:-nats://localhost:4222}"
NATS_ONPREM_SERVER="${NATSSYNC_CLIENT_PORT:-nats://localhost:4223}"
EXIT_APP_TOPIC="${EXIT_APP_TOPIC:-natssync.testing.exitapp}"

# Send exit signal via NATS
if [ `command -v nats` ]; then
  nats pub --server="${NATS_CLOUD_SERVER}" "${EXIT_APP_TOPIC}" '' || true
  nats pub --server="${NATS_ONPREM_SERVER}" "${EXIT_APP_TOPIC}" '' || true
else
	go run apps/natstool.go -u "${NATS_CLOUD_SERVER}" -s "${EXIT_APP_TOPIC}" -m 'hello world' || true
	go run apps/natstool.go -u "${NATS_ONPREM_SERVER}" -s "${EXIT_APP_TOPIC}" -m 'hello world' || true
fi

# Wait until the relevant containers are stopped
source "$(dirname $0)/../single_cluster_test/docker-containers.sh"

limit=12
while [ "${good_to_go}" != "true" ]; do
  good_to_go="true"

  for container in ${SYNCCLIENT} ${SYNCSERVER} ${HTTPPROXYLET} ${HTTPPROXYSERVER}; do
      container_entry=$(docker ps | grep "${container}")
      if [ "${container_entry}" != "" ]; then
          echo "${container} is still running"
          good_to_go="false"
      fi
  done

  if [ ${limit} -le 0 ]; then
    echo "Error: failed to stop all containers: timed out"
    exit 1
  else
    limit=$((limit-1))
  fi

  [ "${good_to_go}" != "true" ] && echo " -- Sleeping 5s before checking again..." && sleep 5
done

echo "Done! All relevant containers have completed gracefully."
