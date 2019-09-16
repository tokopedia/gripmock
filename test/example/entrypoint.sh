#!/usr/bin/env sh

gripmock --stub=example/stubs example/pb/hello.proto &

sleep 5

go run example/client/go/*.go