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
  echo "Done with deployment, waiting 5 seconds for pods to startup..."
  sleep 5
}

getClientPort() {
  namespace=$1
  kubectl -n "$namespace" get service sync-client -o json | jq .spec.ports[0].nodePort
}

register() {
  port=$1
  echo "Registering client on port $port"
  curl -X POST -H 'Content-Type: application/json' -d '{"authToken":"42","locationID":"client1"}' "http://localhost:$port/bridge-client/1/register"
  return $?
}

createOnpremInstance() {
  instanceNum=$1
  namespace="onprem-$instanceNum"
  deployEnvironment "${onpremYamls[*]}" "$namespace"
  port=$(getClientPort "$namespace")

  while true
  do
    register "$port"
    code=$?

    case $code in

    0)
      return 0
      ;;

    52)
      echo "Retry in 5 seconds"
      sleep 5
      ;;

    *)
      echo "Error registering client in namespace $namespace"
      return 1
      ;;
    esac
  done
}

onpremYamls=(
  "redis-pod.yml"
  "redis-service.yml"
  "nats-deployment.yml"
  "nats-service.yml"
  "echoproxylet-deployment.yml"
  "syncclient-scaletest-deployment.yml"
  "syncclient-service-dynport.yml"
)

if [ "$1" ]
then
	count=$1
else
	count=1
fi

i=0

while [ $i -lt $count ]
do
  createOnpremInstance $i &
  ((i++))
done

echo "Instances are deploying in the background"

