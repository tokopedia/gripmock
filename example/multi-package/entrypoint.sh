#!/usr/bin/env sh

# this file is used by .github/workflows/integration-test.yml

# we need to add example/multi-package ar as included import path
# without it protoc could not find the bar/bar.proto
gripmock --stub=example/multi-package/stub --imports=example/multi-package/ \
  example/multi-package/foo.proto example/multi-package/hello.proto \
  example/multi-package/bar/bar.proto &

# wait for generated files to be available and gripmock is up
sleep 2

go run example/multi-package/client/*.go