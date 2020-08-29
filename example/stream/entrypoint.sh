#!/usr/bin/env sh

# this file is used by .github/workflows/integration-test.yml

gripmock --stub=example/stream/stub example/stream &

go run example/stream/client/*.go
