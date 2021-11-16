#!/usr/bin/env bash

#
# Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
#

COVERAGE_DIR="${COVERAGE_DIR:-out}"

[ "${1}" == "" ] && echo "Error: invalid first argument" && echo "Usage example: go_test_app_for_coverage.sh apps/bridge_client_test.go" && exit 1

app_test_filepath="${1}"  # 'apps/bridge_client_test.go'
app_test_filename=$(basename -- "${app_test_filepath}")  # 'bridge_client_test.go'
app_test_name="${app_test_filename%.*}"  # 'bridge_client_test'

go_test_output_filepath="${COVERAGE_DIR}/${app_test_name}_output.txt"
go_coverage_filepath="${COVERAGE_DIR}/${app_test_name}_coverage.out"
junit_report_filepath="${COVERAGE_DIR}/${app_test_name}_report.xml"

echo "${app_test_name}"

go test -v "${app_test_filepath}" -coverprofile="${go_coverage_filepath}" -coverpkg=./pkg/... 2>&1 | tee "${go_test_output_filepath}"
go-junit-report < "${go_test_output_filepath}" > "${junit_report_filepath}"
