#! /usr/bin/bash

#
# Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
#

export CACHE_MGR=mem
export KEYSTORE=file
rm -r -f /out/current/client
mkdir -p /out/current/client
go test -v -coverpkg=github.com/theotw/natssync/pkg/... -coverprofile=/out/current/client/client_coverage.out tests/apps/bridge_client_test.go    2>&1 | tee /out/current/client/client-stdout.txt


go get github.com/t-yuki/gocover-cobertura
go get github.com/wadey/gocovmerge
export PATH=/root/go/bin:$PATH
echo $PATH
ls -l /root/go
gocover-cobertura < /out/current/client/client_coverage.out  > /out/current/client/clientcoverage.xml
mkdir -p /out/previous/
cp -R /out/current/client/* /out/previous/client

