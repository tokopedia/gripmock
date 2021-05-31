#!/usr/bin/env sh

# this file is used by .github/workflows/integration-test.yml

gripmock --stub=example/simple/stub example/simple/simple.proto &

go run example/stub-subfolders/client/*.go