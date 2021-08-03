#!/usr/bin/env sh

# this file is used by .github/workflows/integration-test.yml

gripmock --stub=example/stream/stub example/stream/stream.proto &

# wait for generated files to be available and gripmock is up
sleep 2

go run example/stream/client/*.go