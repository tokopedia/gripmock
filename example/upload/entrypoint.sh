#!/usr/bin/env sh

# this file is used by .github/workflows/integration-test.yml

gripmock &

go run example/upload/client/*.go
