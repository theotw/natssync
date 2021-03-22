#!/bin/bash

REPO="theotw"
NETWORK="natssync-net"

SYNCCLIENT="natssync-client"
SYNCSERVER="natssync-server"
SIMPLEAUTH="simple-reg-auth"
ECHOPROXY="echo-proxylet"
NATSCLOUD="nats-cloud"
MONGOCLOUD="mongo-cloud"
NATSONPREM="nats-onprem"
MONGOONPREM="mongo-onprem"

CONTAINERS=(
  $SYNCCLIENT
  $SYNCSERVER
  $SIMPLEAUTH
  $ECHOPROXY
  $NATSCLOUD
  $MONGOCLOUD
  $NATSONPREM
  $MONGOONPREM
)
