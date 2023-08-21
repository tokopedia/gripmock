ARG BUILD_ARG_GO_VERSION=1.21.0
ARG BUILD_ARG_ALPINE_VERSION=3.18
FROM golang:${BUILD_ARG_GO_VERSION}-alpine${BUILD_ARG_ALPINE_VERSION} AS builder

# install tools (bash, git, protobuf, protoc-gen-go, protoc-grn-go-grpc)
RUN apk -U --no-cache add bash git protobuf &&\
    go install -v github.com/golang/protobuf/protoc-gen-go@latest &&\
    go install -v google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# cloning well-known-types
# only use needed files
RUN git clone --depth=1 https://github.com/protocolbuffers/protobuf.git /protobuf-repo &&\
    mv /protobuf-repo/src/ /protobuf/ &&\
    rm -rf /protobuf-repo

COPY . /go/src/github.com/tokopedia/gripmock

# create necessary dirs and export fix_gopackage.sh
RUN mkdir /proto /stubs &&\
    ln -s /go/src/github.com/tokopedia/gripmock/fix_gopackage.sh /bin/

WORKDIR /go/src/github.com/tokopedia/gripmock/protoc-gen-gripmock

# install generator plugin
RUN go install -v

WORKDIR /go/src/github.com/tokopedia/gripmock/example/simple/client

RUN go get -u all

WORKDIR /go/src/github.com/tokopedia/gripmock

# install gripmock & build example to cache necessary imports
RUN go install -v

# remove pkgs
RUN apk del git

EXPOSE 4770 4771

ENTRYPOINT ["gripmock"]
