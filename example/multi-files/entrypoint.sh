#!/usr/bin/env sh

# this file is used by .github/workflows/integration-test.yml

# Start gripmock in background and capture its output
gripmock --stub=example/multi-files/stub example/multi-files/file1.proto \
  example/multi-files/file2.proto > gripmock.log 2>&1 &

# Wait for gripmock to be ready
wait_for_gripmock.sh

echo "======== RUNNING CLIENT ========="
go run example/multi-files/client/*.go && \
 echo "======== DONE ========="

# kill the server
kill %1