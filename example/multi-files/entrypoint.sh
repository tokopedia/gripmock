#!/usr/bin/env sh

# this file is used by .github/workflows/integration-test.yml

gripmock --stub=example/multi-files/stub example/multi-files/file1.proto \
  example/multi-files/file2.proto &

# wait for generated files to be available and gripmock is up
sleep 2

go run example/multi-files/client/*.go