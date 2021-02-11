#!/bin/bash

set -e

CLOUD_NAMESPACE="cloud"
ONPREM_NAMESPACE="onprem"

kubectl delete namespace "$ONPREM_NAMESPACE" || True
kubectl delete namespace "$CLOUD_NAMESPACE" || True
