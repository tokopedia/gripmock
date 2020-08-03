#!/usr/bin/env sh

# this file is used by .github/workflows/integration-test.yml

gripmock --stub=example/multi-files/stub example/multi-files/file1.proto example/multi-files/file2.proto &

go run example/multi-files/client/*.go