#!/usr/bin/env sh

# this file is used by .github/workflows/integration-test.yml

# we need to add $GOPATH/bar as included import path
# without it protoc could not find the bar.proto
gripmock --stub=example/multi-package/stub --imports=example/multi-package/ example/multi-package/hello.proto &

go run example/multi-package/client/*.go