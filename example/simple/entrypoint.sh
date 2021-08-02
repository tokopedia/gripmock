#!/usr/bin/env sh

# this file is used by .github/workflows/integration-test.yml

gripmock --stub=${GRIPMOCK_DIR}example/simple/stub ${GRIPMOCK_DIR}example/simple/simple.proto &

# wait for generated files to be available and gripmock is up
sleep 2

cat /go/src/grpc/server.go

ls ${GRIPMOCK_DIR}example/simple

go run ${GRIPMOCK_DIR}example/simple/client/*.go