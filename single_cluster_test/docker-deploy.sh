#!/bin/bash

set -e

source "$(dirname $0)/docker-containers.sh"

# URLS
CLOUDKEYSTOREURL="mongodb://$MONGOCLOUD"
CLOUDNATSURL="nats://$NATSCLOUD"
ONPREMKEYSTOREURL="mongodb://$MONGOONPREM"
ONPREMNATSURL="nats://$NATSONPREM"
CLOUD_BRIDGE_URL="http://$SYNCSERVER:8080"

docker network create $NETWORK

# Cloud side
docker run -d --network $NETWORK --name $MONGOCLOUD mongo
docker run -d --network $NETWORK --name $NATSCLOUD -p 4222:4222 nats
docker run -d --network $NETWORK --name $SIMPLEAUTH -e NATS_SERVER_URL=$CLOUDNATSURL $REPO/$SIMPLEAUTH
docker run -d --network $NETWORK --name $SYNCSERVER -p 8080:8080 -e KEYSTORE_URL=$CLOUDKEYSTOREURL -e NATS_SERVER_URL=$CLOUDNATSURL $REPO/$SYNCSERVER

# On-prem side
docker run -d --network $NETWORK --name $MONGOONPREM mongo
docker run -d --network $NETWORK --name $NATSONPREM nats
docker run -d --network $NETWORK --name $ECHOPROXY -e NATS_SERVER_URL=$ONPREMNATSURL $REPO/$ECHOPROXY
docker run -d --network $NETWORK --name $SYNCCLIENT -p 8081:8080 -e KEYSTORE_URL=$ONPREMKEYSTOREURL -e NATS_SERVER_URL=$ONPREMNATSURL -e CLOUD_BRIDGE_URL=$CLOUD_BRIDGE_URL $REPO/$SYNCCLIENT
