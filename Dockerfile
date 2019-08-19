FROM golang:alpine

RUN mkdir /proto

RUN mkdir /stubs

RUN apk -U --no-cache add git protobuf

RUN go get -u -v github.com/golang/protobuf/protoc-gen-go \
	github.com/mitchellh/mapstructure \
	google.golang.org/grpc \
	google.golang.org/grpc/reflection \
	golang.org/x/net/context \
	github.com/go-chi/chi \
	github.com/renstrom/fuzzysearch/fuzzy \
	golang.org/x/tools/imports

RUN go get -u -v github.com/gobuffalo/packr/v2/... \
                 github.com/gobuffalo/packr/v2/packr2

RUN apk del git

RUN mkdir -p /go/src/github.com/tokopedia/gripmock

COPY . /go/src/github.com/tokopedia/gripmock

WORKDIR /go/src/github.com/tokopedia/gripmock/protoc-gen-gripmock

RUN packr2

# install generator plugin
RUN go install -v

RUN packr2 clean

WORKDIR /go/src/github.com/tokopedia/gripmock

# install gripmock
RUN go install -v

RUN rm -rf *

EXPOSE 4770 4771

ENTRYPOINT ["gripmock"]