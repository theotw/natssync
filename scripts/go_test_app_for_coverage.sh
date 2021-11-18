#!/usr/bin/env bash

#
# Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
#

COVERAGE_DIR="${COVERAGE_DIR:-out/}"

[ "${1}" == "" ] && echo "Error: invalid first argument" && echo "Usage example: go_test_app_for_coverage.sh apps/bridge_client_test.go" && exit 1

app_test_filepath="${1}"  # 'apps/bridge_client_test.go'
app_test_filename=$(basename -- "${app_test_filepath}")  # 'bridge_client_test.go'
app_test_name="${TEST_SUITE_NAME}_${app_test_filename%.*}"  # 'mytestsuite_bridge_client_test'

go_test_output_filepath="${COVERAGE_DIR}/${app_test_name}_output.txt"
go_coverage_filepath="${COVERAGE_DIR}/${app_test_name}_coverage.out"
junit_report_filepath="${COVERAGE_DIR}/${app_test_name}_report.xml"

function cleanup_static_files_folders() {
    rm -rf apps/openapi apps/third_party
}

echo "testing app: ${app_test_name}"
date

# Using 'go test' (as we do in this container) will serve from the /build/apps folder instead of /build so we need to move some of
# the static files in there
cleanup_static_files_folders
cp -r openapi/ apps/
mkdir -p apps/third_party && cp -r third_party/swaggerui/ apps/third_party/

go test -v "${app_test_filepath}" -coverprofile="${go_coverage_filepath}" -coverpkg=./pkg/... 2>&1 | tee "${go_test_output_filepath}"
go-junit-report < "${go_test_output_filepath}" > "${junit_report_filepath}"
