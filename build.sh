#!/bin/bash

if [ "$1" = "" ]; then
	echo "Version is empty"
	exit 0
fi

go build ../.

docker buildx build --load -t "tkpd/stubby-gripmock:$1" --platform linux/amd64 .
