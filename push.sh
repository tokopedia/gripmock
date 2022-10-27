#!/bin/bash

if [ "$1" = "" ]; then
	echo "Version is empty"
	exit 0
fi

go build ../.

docker buildx build --push -t "tkpd/gripmock:$1" --platform linux/amd64,linux/arm64 .
