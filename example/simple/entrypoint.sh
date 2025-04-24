#!/usr/bin/env sh

# this file is used by .github/workflows/integration-test.yml

gripmock --stub=example/simple/stub example/simple/simple.proto > gripmock.log 2>&1 &

# Wait for gripmock to be ready
wait_for_gripmock.sh

echo "======== RUNNING CLIENT ========="
go run example/simple/client/*.go && \
 echo "======== DONE ========="