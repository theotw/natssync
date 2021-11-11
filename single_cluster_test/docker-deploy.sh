#!/bin/bash

set -ex

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
  COVERAGE_DIR="${COVERAGE_DIR:-out/coverage}"
  VOLUME_MOUNT_ARG="-v $(pwd)/${COVERAGE_DIR}:/build/${COVERAGE_DIR}"
  NATSSYNC_TEST_IMAGE="${REPO}/natssync-tests:${IMAGE_TAG}"

  SYNCSERVER_FULL_IMG="${NATSSYNC_TEST_IMAGE}"
  SYNCCLIENT_FULL_IMG="${NATSSYNC_TEST_IMAGE}"
  HTTPPROXYSERVER_FULL_IMG="${NATSSYNC_TEST_IMAGE}"
  HTTPPROXYLET_FULL_IMG="${NATSSYNC_TEST_IMAGE}"

  SYNCSERVER_ARGS="-v apps/bridge_server_test.go -coverprofile=${COVERAGE_DIR}/bridge_server_coverage.out -coverpkg=./pkg/..."
  SYNCCLIENT_ARGS="-v apps/bridge_client_test.go -coverprofile=${COVERAGE_DIR}/bridge_client_coverage.out -coverpkg=./pkg/..."
  HTTPPROXYSERVER_ARGS="-v apps/httpproxy_server_test.go -coverprofile=${COVERAGE_DIR}/httpproxy_server_coverage.out -coverpkg=./pkg/..."
  HTTPPROXYLET_ARGS="-v apps/http_proxylet_test.go -coverprofile=${COVERAGE_DIR}/http_proxylet_coverage.out -coverpkg=./pkg/..."
else
  SYNCCLIENT_FULL_IMG="${REPO}/${SYNCCLIENT}:${IMAGE_TAG}"
  SYNCSERVER_FULL_IMG="${REPO}/${SYNCSERVER}:${IMAGE_TAG}"
  HTTPPROXYSERVER_FULL_IMG="${REPO}/${HTTPPROXYSERVER}:${IMAGE_TAG}"
  HTTPPROXYLET_FULL_IMG="${REPO}/${HTTPPROXYLET}:${IMAGE_TAG}"
fi

# Cloud side
docker run -d --network $NETWORK $VOLUME_MOUNT_ARG --name $MONGOCLOUD mongo
docker run -d --network $NETWORK $VOLUME_MOUNT_ARG --name $NATSCLOUD -p 4222:4222 nats
docker run -d --network $NETWORK $VOLUME_MOUNT_ARG --name $SIMPLEAUTH -e LOG_LEVEL=trace -e NATS_SERVER_URL=$CLOUDNATSURL $REPO/$SIMPLEAUTH:$IMAGE_TAG
docker run -d --network $NETWORK $VOLUME_MOUNT_ARG --name $SYNCSERVER -p 8080:8080 -e LOG_LEVEL=trace -e KEYSTORE_URL=$CLOUDKEYSTOREURL -e NATS_SERVER_URL=$CLOUDNATSURL $SYNCSERVER_FULL_IMG $SYNCSERVER_ARGS
docker run -d --network $NETWORK $VOLUME_MOUNT_ARG --name $HTTPPROXYSERVER -p 8082:8080 -e NATS_SERVER_URL=$CLOUDNATSURL $HTTPPROXYSERVER_FULL_IMG $HTTPPROXYSERVER_ARGS

# On-prem side
docker run -d --network $NETWORK $VOLUME_MOUNT_ARG --name $MONGOONPREM mongo
docker run -d --network $NETWORK $VOLUME_MOUNT_ARG --name $NATSONPREM nats
docker run -d --network $NETWORK $VOLUME_MOUNT_ARG --name $ECHOPROXY -e LOG_LEVEL=trace -e NATS_SERVER_URL=$ONPREMNATSURL $REPO/$ECHOPROXY:$IMAGE_TAG
docker run -d --network $NETWORK $VOLUME_MOUNT_ARG --name $SYNCCLIENT -p 8083:8080 -e LOG_LEVEL=trace -e KEYSTORE_URL=$ONPREMKEYSTOREURL -e NATS_SERVER_URL=$ONPREMNATSURL -e CLOUD_BRIDGE_URL=$CLOUD_BRIDGE_URL $SYNCCLIENT_FULL_IMG $SYNCCLIENT_ARGS
docker run -d --network $NETWORK $VOLUME_MOUNT_ARG --name $HTTPPROXYLET -e DEFAULT_LOCATION_ID="*" -e NATS_SERVER_URL=$ONPREMNATSURL $HTTPPROXYLET_FULL_IMG $HTTPPROXYLET_ARGS
docker run -d --network $NETWORK $VOLUME_MOUNT_ARG --name $NGINXTEST $REPO/$NGINXTEST:$IMAGE_TAG
