#!/usr/bin/env sh

# this file is used by .github/workflows/integration-test.yml

# we need to add example/multi-package ar as included import path
# without it protoc could not find the foo.proto & hello.proto
gripmock --stub=example/multi-package/stub \
  example/multi-package/ > gripmock.log 2>&1 &

# Wait for gripmock to be ready
wait_for_gripmock.sh

echo "======== RUNNING CLIENT ========="
go run example/multi-package/client/*.go && \
 echo "======== DONE ========="