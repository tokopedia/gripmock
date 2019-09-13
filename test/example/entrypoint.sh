#!/usr/bin/env sh

gripmock --stub=example/stubs example/pb/hello.proto &

go run example/client/go/*.go