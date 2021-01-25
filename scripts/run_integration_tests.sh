#! /usr/bin/bash

#
# Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
#

export CACHE_MGR=mem
export KEYSTORE=file
echo 'test version 4'
rm -r -f /out/current/integration
mkdir -p /out/current/integration
go test -v tests/integration/server_api_test.go   2>&1 | tee /out/current/integration/server_api_test-stdout.txt
mkdir -p /out/previous/integration
cp -R /out/current/* /out/previous/

