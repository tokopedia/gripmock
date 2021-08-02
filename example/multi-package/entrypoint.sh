#!/usr/bin/env sh

# this file is used by .github/workflows/integration-test.yml

# we need to add example/multi-package ar as included import path
# without it protoc could not find the bar/bar.proto
gripmock --stub=${GRIPMOCK_DIR}example/multi-package/stub --imports=${GRIPMOCK_DIR}example/multi-package/ \
  ${GRIPMOCK_DIR}example/multi-package/foo.proto ${GRIPMOCK_DIR}example/multi-package/hello.proto \
  ${GRIPMOCK_DIR}example/multi-package/bar/bar.proto &

# wait for generated files to be available and gripmock is up
sleep 2

go run ${GRIPMOCK_DIR}example/multi-package/client/*.go