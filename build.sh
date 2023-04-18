#!/bin/bash

if [ "$1" = "" ]; then
	echo "Version is empty"
	exit 0
fi

go mod vendor
(cd protoc-gen-gripmock && go mod vendor)

rm -rf protogen
mkdir -p protogen 

docker buildx build --progress=plain -t "tkpd/gripmock:$1" .
