#!/usr/bin/env sh

# this file is used by .github/workflows/integration-test.yml

gripmock --stub=example/simple-with-gzip/stub example/simple-with-gzip/simple-with-gzip.proto &

# wait for generated files to be available and gripmock is up
sleep 2

go run example/simple-with-gzip/client/*.go
