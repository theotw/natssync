#!/bin/bash

set -e

source "$(dirname $0)/docker-containers.sh"

# URLS
CLOUDKEYSTOREURL="mongodb://$MONGOCLOUD"
CLOUDNATSURL="nats://$NATSCLOUD"
ONPREMKEYSTOREURL="mongodb://$MONGOONPREM"
ONPREMNATSURL="nats://$NATSONPREM"
CLOUD_BRIDGE_URL="http://$SYNCSERVER:8080"

IMAGE_TAG="latest"

if [ "$1" ]
then
  IMAGE_TAG=$1
fi

echo "IMAGE_TAG=$IMAGE_TAG"

docker network create $NETWORK

if [ "${TEST_MODE_ENABLED}" = "true" ]; then
  NATSSYNC_TEST_IMAGE="natssync-tests"
  SYNCCLIENT="${NATSSYNC_TEST_IMAGE}"
  SYNCSERVER="${NATSSYNC_TEST_IMAGE}"
fi

# Cloud side
docker run -d --network $NETWORK --name $MONGOCLOUD mongo
docker run -d --network $NETWORK --name $NATSCLOUD -p 4222:4222 nats
docker run -d --network $NETWORK --name $SIMPLEAUTH -e LOG_LEVEL=trace -e NATS_SERVER_URL=$CLOUDNATSURL $REPO/$SIMPLEAUTH:$IMAGE_TAG
docker run -d --network $NETWORK --name $SYNCSERVER -p 8080:8080 -e LOG_LEVEL=trace -e KEYSTORE_URL=$CLOUDKEYSTOREURL -e NATS_SERVER_URL=$CLOUDNATSURL $REPO/$SYNCSERVER:$IMAGE_TAG

# On-prem side
docker run -d --network $NETWORK --name $MONGOONPREM mongo
docker run -d --network $NETWORK --name $NATSONPREM nats
docker run -d --network $NETWORK --name $ECHOPROXY -e LOG_LEVEL=trace -e NATS_SERVER_URL=$ONPREMNATSURL $REPO/$ECHOPROXY:$IMAGE_TAG
docker run -d --network $NETWORK --name $SYNCCLIENT -p 8081:8080 -e LOG_LEVEL=trace -e KEYSTORE_URL=$ONPREMKEYSTOREURL -e NATS_SERVER_URL=$ONPREMNATSURL -e CLOUD_BRIDGE_URL=$CLOUD_BRIDGE_URL $REPO/$SYNCCLIENT:$IMAGE_TAG
