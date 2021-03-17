#!/bin/bash

set -e

CLOUD_NAMESPACE="cloud"
ONPREM_NAMESPACE="onprem"

deleteAllOfKind() {
	kind=$1
	namespace=$2
	kubectl delete --all "$kind" -n "$namespace"
}

deleteNamespace() {
	namespace=$1
	deleteAllOfKind deployments "$namespace"
	deleteAllOfKind pods "$namespace"
	deleteAllOfKind services "$namespace"
	kubectl delete namespace "$namespace" || True
}

deleteNamespace "$ONPREM_NAMESPACE"
deleteNamespace "$CLOUD_NAMESPACE"

