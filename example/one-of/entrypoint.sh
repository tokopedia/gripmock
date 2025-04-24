#!/usr/bin/env sh

# this file is used by .github/workflows/integration-test.yml

# start the server
gripmock --stub=example/one-of/stub example/one-of/oneof.proto > gripmock.log 2>&1 &

# Wait for gripmock to be ready
wait_for_gripmock.sh

echo "======== RUNNING CLIENT ========="
go run example/one-of/client/*.go && \
 echo "======== DONE ========="

# kill the server
kill %1