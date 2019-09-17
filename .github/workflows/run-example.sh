#!/usr/bin/env sh

gripmock --stub=example/$1/stub example/$1/$1.proto &

go run example/$1/client/*.go