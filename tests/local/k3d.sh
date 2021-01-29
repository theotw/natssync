#!/bin/bash

kubectl config use-context k3d-nats 2> /dev/null 1> /dev/null

if [ $? != 0 ]; then
  #k3d cluster create nats -i rancher/k3s:v1.18.15-k3s1 --api-port 127.0.0.1:6443 -p 80:80@loadbalancer -p 443:443@loadbalancer
  k3d cluster create nats -i rancher/k3s:v1.19.7-k3s1 --config k3d.yaml --registry-create
else
  echo "---------"
  echo "Recreating NATS k3d cluster..."
  echo "---------"
  k3d cluster delete nats
  k3d cluster create nats -i rancher/k3s:v1.19.7-k3s1 --config k3d.yaml --registry-create
fi
