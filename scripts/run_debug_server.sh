#! /usr/bin/bash

#
# Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
#

export CACHE_MGR=mem
export KEYSTORE=file

#go test -v -coverpkg=github.com/theotw/natssync/pkg/... -coverprofile=out/server_coverage.out tests/apps/bridge_server_test.go  2>&1 >out/server-stdout.txt
rm -r -f /out/current/server
mkdir -p /out/current/server
go test -v -coverpkg=github.com/theotw/natssync/pkg/... -coverprofile=/out/current/server_coverage.out tests/apps/bridge_server_test.go   2>&1 | tee /out/current/server-stdout.txt

mkdir -p /out/previous/server
cp /out/current/server/* /out/previous/server
