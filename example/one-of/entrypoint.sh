#!/usr/bin/env sh

# this file is used by .github/workflows/integration-test.yml

gripmock --stub=${GRIPMOCK_DIR}example/one-of/stub ${GRIPMOCK_DIR}example/one-of/oneof.proto &

# wait for generated files to be available and gripmock is up
sleep 2

go run ${GRIPMOCK_DIR}example/one-of/client/*.go