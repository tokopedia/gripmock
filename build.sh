#!/bin/bash

if [ "$1" = "" ]; then
	echo "Version is empty"
	exit 0
fi

go build ../.

docker build -t "quintans/gripmock:$1" .