#!/usr/bin/env sh

# this file is used by .github/workflows/integration-test.yml

gripmock --stub=example/no-gopackage/stub \
  example/no-gopackage/foo.proto example/no-gopackage/hello.proto \
  example/no-gopackage/bar/bar.proto \
  example/no-gopackage/bar/deep/bar.proto &

# wait for generated files to be available and gripmock is up
sleep 2

go run example/no-gopackage/client/*.go