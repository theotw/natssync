# CICD

## Overview

### Intro

This document described the CICD pipeline for this project.

The build is in in github actions defined in .github/workflows/cicd.yml

We do have codeql workflow also described as well as the github security scanning enabled for this project

### Definitions

- Unit tests, also called l1 tests are in line with the code and do not need a running system to execute
- Integration tests, also called l2 require a running system to execute and are defined in tests/integration

## Pipeline

The pipeline has the following phases

- Unit Tests (aka l1) runs all tests in the pkg subdirectory
- Build which builds the docker images
- Integration (aka l2) which deployes the docker images and runs the test is tests/integration
- If successful, we tag images to the latest and push the dated versions and latest to dockerhub

## Adding tests

Adding tests is simple, for unit tests (See definition) add them inline with code per defacto go standards To add in
integration tests add them under tests/integration. Use the go test format / style.

For integration, the following env vars are defined for use in integration tests

* export syncserver_url='http://localhost:8080'
* export syncclient_url='http://localhost:8081'
* export natsserver_url='nats://localhost:4222'
