#!/bin/bash

set -e

CLOUD_NAMESPACE="cloud"
ENT_NAMESPACE="ent"

kubectl delete namespace "$CLOUD_NAMESPACE" || True
kubectl delete namespace "$ENT_NAMESPACE" || True
