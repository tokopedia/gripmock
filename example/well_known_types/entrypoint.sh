#!/usr/bin/env sh

# this file is used by .github/workflows/integration-test.yml

gripmock --stub=example/well_known_types/stub example/well_known_types &

go run example/well_known_types/client/*.go
