#!/usr/bin/env sh

# this file is used by .github/workflows/integration-test.yml

gripmock --stub=${GRIPMOCK_DIR}example/simple/stub ${GRIPMOCK_DIR}example/simple/simple.proto &

pwd

# wait for generated files to be available and gripmock is up
sleep 2

go run ${GRIPMOCK_DIR}example/simple/client/*.go