#!/usr/bin/env sh

# this file is used by .github/workflows/integration-test.yml

gripmock --stub=${GRIPMOCK_DIR}example/well_known_types/stub ${GRIPMOCK_DIR}example/well_known_types/wkt.proto &

# wait for generated files to be available and gripmock is up
sleep 2

go run ${GRIPMOCK_DIR}example/well_known_types/client/*.go