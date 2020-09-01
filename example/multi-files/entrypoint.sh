#!/usr/bin/env sh

# this file is used by .github/workflows/integration-test.yml

gripmock --stub=example/multi-files/stub example/multi-files &

go run example/multi-files/client/*.go
