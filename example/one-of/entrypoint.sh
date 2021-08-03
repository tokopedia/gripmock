#!/usr/bin/env sh

# this file is used by .github/workflows/integration-test.yml

gripmock --stub=example/one-of/stub example/one-of/oneof.proto &

# wait for generated files to be available and gripmock is up
sleep 2

go run example/one-of/client/*.go