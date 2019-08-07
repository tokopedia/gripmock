FROM golang:alpine

RUN mkdir /proto

RUN mkdir /stubs

RUN apk -U --no-cache add git protobuf

RUN go get -u -v github.com/golang/protobuf/protoc-gen-go \
	github.com/mitchellh/mapstructure \
	google.golang.org/grpc \
	google.golang.org/grpc/reflection \
	golang.org/x/net/context \
	github.com/alecthomas/participle \
	github.com/go-chi/chi \
	github.com/renstrom/fuzzysearch/fuzzy

RUN mkdir -p /go/src/github.com/charandas/gripmock

COPY . /go/src/github.com/charandas/gripmock

WORKDIR /go/src/github.com/charandas/gripmock

RUN go build

RUN mv /go/src/github.com/charandas/gripmock/gripmock /usr/bin/gripmock

RUN rm -rf *

EXPOSE 4770 4771

RUN apk del git
