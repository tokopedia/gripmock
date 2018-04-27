FROM golang:alpine

RUN mkdir /proto

RUN apk -U --no-cache add git protobuf

RUN go get -u -v github.com/golang/protobuf/protoc-gen-go \
	github.com/mitchellh/mapstructure \
	google.golang.org/grpc \
	google.golang.org/grpc/reflection \
	golang.org/x/net/context

RUN mkdir -p /go/src/github.com/ahmadmuzakki/grpcmock

COPY . /go/src/github.com/ahmadmuzakki/grpcmock

WORKDIR /go/src/github.com/ahmadmuzakki/grpcmock

RUN go build

RUN mv /go/src/github.com/ahmadmuzakki/grpcmock/grpcmock /usr/bin/grpcmock

RUN rm -rf *

EXPOSE 4770 4771

RUN apk del git
