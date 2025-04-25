#!/usr/bin/env sh

# this file is used by .github/workflows/integration-test.yml

gripmock --stub=example/well_known_types/stub example/well_known_types/wkt.proto > gripmock.log 2>&1 &

# Wait for gripmock to be ready
wait_for_gripmock.sh

echo "======== RUNNING CLIENT ========="
go run example/well_known_types/client/*.go && \
 echo "======== DONE ========="

# kill the server
kill %1