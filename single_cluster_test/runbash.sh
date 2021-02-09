#!/bin/bash
set -e

NAMESPACE=$1

if [ -z "$NAMESPACE" ]
then
  echo "Missing namespace argument"
  exit 1
fi

kubectl run my-shell -n "$NAMESPACE" --rm -i --tty --image ubuntu -- bash
