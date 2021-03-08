#!/bin/bash
set -e

makeNamespace() {
  kubectl create namespace "$1" || echo "Namespace $1 already exists"
}

deployEnvironment() {
  yamls=$1
  namespace=$2

  makeNamespace "$namespace"

  for yaml in ${yamls[*]}
  do
    echo "Applying $yaml..."
    kubectl apply -n "$namespace" -f "$yaml"
  done
}

CLOUD_NAMESPACE="cloud"
ONPREM_NAMESPACE="onprem"

cloudYamls=(
  "nats-deployment.yml"
  "nats-nodeport-service.yml"
  "redis-pod.yml"
  "redis-service.yml"
  "simple-reg-deployment.yml"
  "syncserver-deployment.yml"
  "syncserver-service.yml"
  "mongo-pod.yml"
  "mongo-service.yml"
)

onpremYamls=(
  "redis-pod.yml"
  "redis-service.yml"
  "nats-deployment.yml"
  "nats-service.yml"
  "echoproxylet-deployment.yml"
  "syncclient-deployment.yml"
  "syncclient-service.yml"
 )

deployEnvironment "${cloudYamls[*]}" "$CLOUD_NAMESPACE"
deployEnvironment "${onpremYamls[*]}" "$ONPREM_NAMESPACE"
