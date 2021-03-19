#!/bin/bash

set -e

source "$(dirname $0)/docker-containers.sh"

if [ -z "$1" ]
then
  echo "Please provide a log directory name"
  exit 1
fi

logsDir=$1

mkdir -p "$logsDir"

for container in ${CONTAINERS[*]}
do
  docker logs "$container" >> "${logsDir}/${container}.txt" 2>&1
done
