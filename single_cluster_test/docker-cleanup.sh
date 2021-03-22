#!/bin/bash

source "$(dirname $0)/docker-containers.sh"

echo "Stopping containers"
docker stop ${CONTAINERS[*]}
echo "Removing Containers"
docker rm ${CONTAINERS[*]}
echo "Removing network"
docker network rm $NETWORK

exit 0
