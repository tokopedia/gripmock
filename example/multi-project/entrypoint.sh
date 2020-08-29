#!/usr/bin/env sh

# this file is used by .github/workflows/integration-test.yml

gripmock --stub=example/multi-project/stub -isd example/multi-project/proto &

go run example/multi-project/client/*.go
