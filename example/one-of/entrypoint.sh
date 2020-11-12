#!/usr/bin/env sh

# this file is used by .github/workflows/integration-test.yml

gripmock --stub=example/one-of/stub example/one-of/oneof.proto &

go run example/one-of/client/*.go