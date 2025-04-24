#!/usr/bin/env sh

# this file is used by .github/workflows/integration-test.yml

# start the server
gripmock --stub=example/stub-subfolders/stub example/stub-subfolders/stub-subfolders.proto > gripmock.log 2>&1 &

# Wait for gripmock to be ready
wait_for_gripmock.sh

echo "======== RUNNING CLIENT ========="
go run example/stub-subfolders/client/*.go && \
 echo "======== DONE ========="

# kill the server
kill %1