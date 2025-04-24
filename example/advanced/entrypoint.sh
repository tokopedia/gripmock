#!/usr/bin/env sh

# this file is used by .github/workflows/integration-test.yml

gripmock --stub=example/advanced/stub example/advanced/advanced.proto > gripmock.log 2>&1 &

# Wait for gripmock to be ready
wait_for_gripmock.sh

echo "======== RUNNING CLIENT ========="
go run example/advanced/client/*.go && \
 echo "======== DONE ========="

# kill the server
kill %1 