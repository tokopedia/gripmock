#!/usr/bin/env sh

# this file is used by .github/workflows/integration-test.yml

gripmock --stub=example/multi-project/stub --imports=example/multi-project/proto/prj-bar,example/multi-project/proto/prj-foo example/multi-project/proto &

go run example/multi-project/client/*.go
