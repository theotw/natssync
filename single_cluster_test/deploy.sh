#!/bin/bash
set -e

makeNamespace() {
  kubectl create namespace "$1" || echo "Namespace $namespace already exists"
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
ENT_NAMESPACE="ent"

cloudYamls=(
  "nats-deployment.yml"
  "nats-nodeport-service.yml"
  "redis-pod.yml"
  "redis-service.yml"
  "simple-reg-deployment.yml"
  "syncserver-deployment.yml"
  "syncserver-service.yml"
)

entYamls=(
  "redis-pod.yml"
  "redis-service.yml"
  "nats-deployment.yml"
  "nats-service.yml"
  "echoproxylet-deployment.yml"
  "syncclient-deployment.yml"
 )

deployEnvironment "${cloudYamls[*]}" "$CLOUD_NAMESPACE"
deployEnvironment "${entYamls[*]}" "$ENT_NAMESPACE"
