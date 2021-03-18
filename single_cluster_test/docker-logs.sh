#!/bin/bash

set -e

source "$(dirname $0)/docker-containers.sh"

if [ -z "$1" ]
then
  echo "Please provide a log file name"
  exit 1
fi

logfile=$1

for container in $CONTAINERS
do
  echo "==========$container==========" >> "$logfile"
  docker logs "$container" >> "$logfile"
done
