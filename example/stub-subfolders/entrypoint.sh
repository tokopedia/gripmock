#!/usr/bin/env sh

# this file is used by .github/workflows/integration-test.yml

gripmock --stub=example/stub-subfolders/stub example/stub-subfolders/stub-subfolders.proto &

go run example/stub-subfolders/client/*.go