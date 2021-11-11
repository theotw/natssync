#!/usr/bin/env bash

curl -X GET http://localhost:8083/bridge-client/kill
curl -X GET http://localhost:8080/bridge-server/kill
