#!/usr/bin/env sh

# this file is used by .github/workflows/integration-test.yml

gripmock --stub=example/stream/stub example/stream/stream.proto &

go run example/stream/client/*.go