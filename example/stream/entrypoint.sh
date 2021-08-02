#!/usr/bin/env sh

# this file is used by .github/workflows/integration-test.yml

gripmock --stub=${GRIPMOCK_DIR}example/stream/stub ${GRIPMOCK_DIR}example/stream/stream.proto &

# wait for generated files to be available and gripmock is up
sleep 2

go run ${GRIPMOCK_DIR}example/stream/client/*.go