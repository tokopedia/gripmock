#!/usr/bin/env sh

# this file is used by .github/workflows/integration-test.yml

gripmock --stub=${GRIPMOCK_DIR}example/multi-files/stub ${GRIPMOCK_DIR}example/multi-files/file1.proto \
  ${GRIPMOCK_DIR}example/multi-files/file2.proto &

# wait for generated files to be available and gripmock is up
sleep 2

go run ${GRIPMOCK_DIR}example/multi-files/client/*.go