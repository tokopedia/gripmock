#!/usr/bin/env sh

gripmock --grpc-listen='0.0.0.0' --stub=example/stubs example/pb/hello.proto

#go run example/client/go/*.go